package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoDBClusterReconciler) reconcileConfigServerReplicaSet(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster) error {
	var members []string
	for i := 0; i < mongo.Spec.ConfigServerCount; i++ {
		host := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:27019", "config", i, mongo.Name+"-configsvr", mongo.Namespace)
		members = append(members, host)

		return waitForReplicaSetHealthy(host, 5*time.Second)
	}

	host, err := findReachableHost(ctx, members)
	if err != nil {
		return err
	}

	return InitReplicaSet(ctx, "configReplSet", host, true, members...)
}

func (r *MongoDBClusterReconciler) reconcileReplicaSet(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster, shardID int) error {
	var members []string
	for i := 0; i < mongo.Spec.ReplicaSetSize; i++ {
		host := fmt.Sprintf("%s%d-%d.%s%d.%s.svc.cluster.local:27018", "mongo-shard", shardID, i, mongo.Name+"-shard", shardID, mongo.Namespace)
		waitForReplicaSetHealthy(host, 3*time.Second)
		members = append(members, host)
	}

	host, err := findReachableHost(ctx, members)
	if err != nil {
		return err
	}

	replicaSetName := fmt.Sprintf("%s%d", "mongo-shard", shardID)

	return InitReplicaSet(ctx, replicaSetName, host, false, members...)
}

func (r *MongoDBClusterReconciler) waitForAllReplicas(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster, shardID int) error {
	replicaCount := mongo.Spec.ReplicaSetSize
	var failedHosts []string

	for i := 0; i < replicaCount; i++ {
		host := fmt.Sprintf("mongo-shard%d-%d.%s%d.%s.svc.cluster.local:27018", shardID, i, mongo.Name+"-shard", shardID, mongo.Namespace)

		success := false
		for {
			select {
			case <-ctx.Done():
				failedHosts = append(failedHosts, host)
				break
			default:
				err := waitForReplicaSetHealthy(host, 1*time.Second)
				if err == nil {
					success = true
					break
				}
				time.Sleep(1 * time.Second)
			}
			if success {
				break
			}
		}
	}

	return nil
}

func findReachableHost(ctx context.Context, members []string) (string, error) {
	for _, host := range members {
		connStr := fmt.Sprintf("mongodb://%s", host)
		clientOpts := options.Client().ApplyURI(connStr).SetDirect(true).SetServerSelectionTimeout(2 * time.Second)
		client, err := mongo.Connect(ctx, clientOpts)
		if err == nil {
			if err = client.Ping(ctx, nil); err == nil {
				_ = client.Disconnect(ctx)
				return host, nil
			}
			_ = client.Disconnect(ctx)
		}
	}
	return "", fmt.Errorf("no reachable MongoDB host found")
}

// InitReplicaSet initializes the MongoDB replica set.
func InitReplicaSet(ctx context.Context, replicaSetName, host string, configServer bool, members ...string) error {
	if len(members) == 0 {
		return fmt.Errorf("No host for replication MongoDB: %w")
	}
	// Connect to the first member (assumed to become primary)
	connStr := fmt.Sprintf("mongodb://%s", host)
	clientOpts := options.Client().ApplyURI(connStr).SetDirect(true)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	defer client.Disconnect(ctx)

	adminDB := client.Database("admin")

	var result bson.M
	err = adminDB.RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&result)
	if err == nil {
		return nil
	}

	// memberList := bson.A{}
	// for i, m := range members {
	// 	memberList = append(memberList, bson.D{
	// 		{Key: "_id", Value: i},
	// 		{Key: "host", Value: m},
	// 	})
	// }
	//
	// config := bson.D{
	// 	{Key: "_id", Value: replicaSetName},
	// 	{Key: "configsvr", Value: true},
	// 	{Key: "members", Value: memberList},
	// }

	memberList := bson.A{}
	for i, m := range members {
		memberList = append(memberList, bson.D{
			{Key: "_id", Value: i},
			{Key: "host", Value: m},
		})
	}

	config := bson.D{
		{Key: "_id", Value: replicaSetName},
		{Key: "members", Value: memberList},
	}

	if configServer {
		config = append(config[:1], bson.D{{Key: "configsvr", Value: true}}[0])
		config = append(config, bson.D{{Key: "members", Value: memberList}}[0])
	}

	initCmd := bson.D{{Key: "replSetInitiate", Value: config}}

	if err := adminDB.RunCommand(ctx, initCmd).Err(); err != nil {
		return fmt.Errorf("replSetInitiate failed: %w", err)
	}

	return waitForPrimary(ctx, adminDB)
}

func waitForPrimary(ctx context.Context, adminDB *mongo.Database) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for replica set to become primary")
		case <-ticker.C:
			var status bson.M
			err := adminDB.RunCommand(ctx, bson.D{{Key: "replSetGetStatus", Value: 1}}).Decode(&status)
			if err != nil {
				continue
			}

			members, ok := status["members"].(bson.A)
			if !ok {
				continue
			}

			for _, m := range members {
				memberMap := m.(bson.M)
				if stateStr, ok := memberMap["stateStr"].(string); ok && stateStr == "PRIMARY" {
					return nil
				}
			}
		}
	}
}

func waitForReplicaSetHealthy(uri string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI("mongodb://" + uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer client.Disconnect(ctx)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return errors.New("timeout waiting for MongoDB to be reachable")

		case <-ticker.C:
			if err := client.Ping(ctx, nil); err == nil {
				return nil // success
			}
		}
	}
}
