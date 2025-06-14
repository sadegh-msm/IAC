# Full-Stack DevOps Project

## Overview

This project is a comprehensive full-stack DevOps implementation that spans infrastructure provisioning, Kubernetes cluster setup, CI/CD pipelines, application development, and custom operator creation. It follows a structured four-step approach to build, deploy, and manage a highly available, production-ready platform using modern tools and practices.

---

## Steps & Technical Breakdown

### **Step 1: Linux, Ansible, and Infrastructure Provisioning**

#### Tasks:

* Provision 2 VMs on **Arvan Cloud** using **Terraform**.
* Use **Ansible** to automate:

  * Passwordless SSH setup for a non-root user.
  * Disable `su` access; enforce `sudo` for privilege escalation.
  * Install and configure **Bind9** as a DNS server.
  * Deploy **GitLab**, **GitLab Container Registry**, and **GitLab Runner** using Docker.
  * Auto-register GitLab Runner with the GitLab instance.

---

### **Step 2: Kubernetes, Monitoring, and Backup**

#### Tasks:

* Extend the Ansible playbook to provision **3 new VMs** (1 control plane, 2 workers).
* Install and configure:

  * **kubeadm**, **CRI-O**, **Calico**, and **Longhorn**.
  * High Availability (HA) Kubernetes cluster with `--control-plane-endpoint`.
* Deploy backup & monitoring tools:

  * **Velero** for scheduled backups to **Arvan Object Storage**.
  * **VictoriaMetrics** to scrape `kube-system` metrics.
  * **Grafana** dashboard for CPU, memory, and pod status visualization.

---

### **Step 3: Python/Golang, Redis, MongoDB, and CI/CD**

#### Tasks:

* Develop a **URL Shortener** using:

  * **FastAPI** for backend logic.
  * **MongoDB** (with TTL index) for storage.
  * **Redis** for caching hot URLs.
* Kubernetes Deployment:

  * Use **Helm charts** for MongoDB and Redis.
  * Deploy the application via Helm chart or **Kustomize overlay**.
* CI/CD:

  * Build and push Docker image to **GitLab Registry**.
  * Set up auto-deployment via GitLab CI on `git push`.

---

### **Step 4: Kubernetes Operator Development**

#### Tasks:

* Build a **MongoDB Operator** using **Kubebuilder**.
* Define a custom resource: `MongoDBCluster` with fields for:

  * Sharding
  * Replication
  * Backup configuration
* Implement controller logic to:

  * Deploy MongoDB StatefulSets.
  * Set up replication and sharding automatically.
  * Trigger backups via **Velero** or `mongodump` to Arvan storage.
* Test:

  * Deploy the Operator to the cluster.
  * Create a `MongoDBCluster` resource and verify proper reconciliation.

---

## Tech Stack

* **Infrastructure:** Terraform, Ansible, Docker, Bind9
* **Kubernetes:** kubeadm, CRI-O, Calico, Longhorn
* **Monitoring & Backup:** VictoriaMetrics, Grafana, Velero
* **Backend:** FastAPI (Python) or Go
* **Data:** MongoDB, Redis
* **DevOps Tools:** GitLab, GitLab Runner, Helm, Kustomize
* **Operator Framework:** Kubebuilder

---

## Requirements

* Arvan Cloud Account
* Ansible & Terraform installed
* Docker installed on target hosts
* Kubernetes CLI tools: `kubeadm`, `kubectl`, `crictl`
* GitLab instance access

