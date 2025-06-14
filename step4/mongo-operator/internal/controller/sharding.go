package controller

import (
	"context"
	"fmt"
	"strings"
	"time"

	databasev1alpha1 "github.com/sadegh-msm/mongodb-operator/api/v1alpha1"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *MongoDBClusterReconciler) reconcileEnbaleShardingMongos(ctx context.Context, mongo *databasev1alpha1.MongoDBCluster, shardID int) error {
	var shards []Shard

	replicaSetName := fmt.Sprintf("mongo-shard%d", shardID)
	serviceName := fmt.Sprintf("%s-shard%d", mongo.Name, shardID) // e.g., mongodbcluster-sample-shard0
	var members []string

	for j := 0; j < mongo.Spec.ReplicaSetSize; j++ {
		podDNS := fmt.Sprintf("%s-%d.%s.%s.svc.cluster.local:%d", replicaSetName, j, serviceName, mongo.Namespace, 27018)
		members = append(members, podDNS)
	}

	shards = append(shards, Shard{
		ReplicaSetName: replicaSetName,
		Members:        members,
	})

	cfg := ShardConfig{
		MongosURI:       fmt.Sprintf("mongodb://%s-%s.%s.svc.cluster.local:27017", mongo.Name, "mongos", mongo.Namespace),
		Shards:          shards,
		EnableDB:        mongo.Spec.Sharding.Database,
		ShardCollection: mongo.Spec.Sharding.Collections,
		ShardKey:        bson.D{{Key: "_id", Value: "hashed"}},
	}

	return InitSharding(ctx, cfg)
}

// Shard represents a shard replica set to be added to the cluster.
type Shard struct {
	ReplicaSetName string   // e.g., "shard0"
	Members        []string // e.g., ["shard0-0.mongodb.default.svc.cluster.local:27017", ...]
}

// ShardConfig holds all options for initializing sharding.
type ShardConfig struct {
	MongosURI       string  // mongos URI, e.g., mongodb://mongos.default.svc.cluster.local:27017
	Shards          []Shard // list of shard replica sets
	EnableDB        string  // optional: name of database to enable sharding on
	ShardCollection string  // optional: fully qualified collection name, e.g., "mydb.mycollection"
	ShardKey        bson.D  // optional: shard key, e.g., bson.D{{"userId", 1}}
}

// InitSharding connects to mongos and adds shards + optionally enables sharding.
func InitSharding(ctx context.Context, cfg ShardConfig) error {
	clientOpts := options.Client().ApplyURI(cfg.MongosURI)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return fmt.Errorf("failed to connect to mongos: %w", err)
	}
	defer client.Disconnect(ctx)

	adminDB := client.Database("admin")

	for _, shard := range cfg.Shards {
		if len(shard.Members) == 0 {
			continue
		}

		// if err := waitForPrimaryShard(ctx, shard); err != nil {
		// 	return fmt.Errorf("primary not found for shard %s: %w", shard.ReplicaSetName, err)
		// }

		shardURI := fmt.Sprintf("%s/%s", shard.ReplicaSetName, strings.Join(shard.Members, ","))

		cmd := bson.D{{Key: "addShard", Value: shardURI}}
		if err := adminDB.RunCommand(ctx, cmd).Err(); err != nil && !isAlreadyExistsError(err) {
			return fmt.Errorf("failed to add shard %s: %w", shard.ReplicaSetName, err)
		}
	}

	if cfg.EnableDB != "" {
		cmd := bson.D{{Key: "enableSharding", Value: cfg.EnableDB}}
		if err := adminDB.RunCommand(ctx, cmd).Err(); err != nil && !isAlreadyEnabledError(err) {
			return fmt.Errorf("failed to enable sharding on db %s: %w", cfg.EnableDB, err)
		}
	}

	if cfg.ShardCollection != "" && len(cfg.ShardKey) > 0 {
		cmd := bson.D{
			{Key: "shardCollection", Value: cfg.ShardCollection},
			{Key: "key", Value: cfg.ShardKey},
		}
		if err := adminDB.RunCommand(ctx, cmd).Err(); err != nil {
			return fmt.Errorf("failed to shard collection %s: %w", cfg.ShardCollection, err)
		}
	}

	return nil
}

func isAlreadyExistsError(err error) bool {
	return err != nil && (contains(err.Error(), "already exists") || contains(err.Error(), "duplicate key"))
}

func isAlreadyEnabledError(err error) bool {
	return err != nil && contains(err.Error(), "already enabled")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr)))
}

func waitForPrimaryShard(ctx context.Context, shard Shard) error {
	const timeout = time.Minute
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		for _, member := range shard.Members {
			memberURI := fmt.Sprintf("mongodb://%s/?replicaSet=%s", member, shard.ReplicaSetName)
			clientOpts := options.Client().ApplyURI(memberURI).SetDirect(true) // direct connection to this member
			client, err := mongo.Connect(ctx, clientOpts)
			if err != nil {
				continue
			}

			defer client.Disconnect(ctx)

			var result bson.M
			err = client.Database("admin").RunCommand(ctx, bson.D{{Key: "isMaster", Value: 1}}).Decode(&result)
			if err == nil {
				if primary, ok := result["ismaster"].(bool); ok && primary {
					// primary found
					return nil
				}
			}
		}

		// Wait 2 seconds before retrying
		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for primary on shard %s", shard.ReplicaSetName)
}
