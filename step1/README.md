# Step 1 - Part 1: VM Provisioning on Arvan Cloud with Terraform

## Overview

This Terraform configuration automates the provisioning of **2 virtual machines** on **Arvan Cloud**, setting the foundation for further DevOps automation via Ansible.

It dynamically fetches the latest compatible Ubuntu image, selects the desired VM plan, and creates instances using secure SSH access.

---

## Directory Structure

```
step1-terraform/
├── main.tf
├── variables.tf
├── terraform.tfvars
```

---

## Features

* Provisions **2 VMs** in a specified Arvan Cloud region.
* Dynamically selects OS image and VM flavor by name.
* Uses **SSH key authentication** (key name: `macbook`).
* Attaches the default **security group** to each VM.
* Configurable via input variables.

---

## Configuration

### Required Variables

| Name                 | Description                             | Default        |
| -------------------- | --------------------------------------- | -------------- |
| `api_key`            | Arvan Cloud API key                     | **(Required)** |
| `region`             | Region for VM deployment                | `ir-thr-ba1`   |
| `chosen_distro_name` | Name of the OS distro (e.g., ubuntu)    | `ubuntu`       |
| `chosen_name`        | Specific OS version                     | `24.04`        |
| `chosen_plan_id`     | Plan ID for VM flavor (e.g., g2-12-4-0) | `g2-12-4-0`    |

You can override these values in a `terraform.tfvars` file or via CLI input.

---

## Usage

1. **Install Terraform** (>= 1.0 recommended)
2. **Initialize the project**:

   ```bash
   terraform init
   ```
3. **Preview the plan**:

   ```bash
   terraform plan
   ```
4. **Apply the configuration**:

   ```bash
   terraform apply
   ```

---

## Resources Created

* `2x arvan_abrak` VM instances
* Each with:

  * 25 GB disk
  * SSH access via `macbook` SSH key
  * Default security group

---

## Teardown

To destroy the resources:

```bash
terraform destroy
```

---

## Notes

* The `ssh_key_name` is currently hardcoded as `"macbook"`. You can parameterize it if needed.
* Ensure that the SSH key named `"macbook"` is already added to your Arvan account.


---

# Step 1 – Part 2: Ansible Playbook – Server Configuration

## Overview

This Ansible playbook configures the two virtual machines provisioned with Terraform to be production-ready. It includes:

* ✅ Passwordless SSH configuration for a non-root user
* ✅ System hardening by disabling `su` and enforcing `sudo`
* ✅ DNS service installation using `Bind9`
* ✅ GitLab stack deployment (GitLab, Registry, and Runner in Docker)

---

## Playbook Structure

```yaml
- name: Configure Servers
  hosts: all
  become: true
  roles:
    - roles/ssh_passwordless
    - roles/hardening
    - roles/bind9
    - roles/gitlab
```

This main playbook calls four logical roles, each explained below.

---

## Role Breakdown

### 1. `ssh_passwordless`: Configure Passwordless SSH

This role:

* Creates a user (e.g., `ubuntu`)
* Sets up `~/.ssh` directory with proper permissions
* Installs the public SSH key from the control machine (`public_key_file`)

This ensures seamless Ansible operations and secure login without using a password.

To use this role set Tag to `ssh_passwordless`

---

### 2. `hardening`: Enforce Sudo and Restrict `su`

Security best practices implemented:

* Ensures `sudo` is installed
* Adds the user to the `sudo` group
* Disables `su` access to anyone outside the `sudo` group (via PAM)
* Configures the `sudoers` file to require password for escalations

**Key File Updated:**

* `/etc/pam.d/su`: Disables unrestricted `su`
* `/etc/sudoers.d/90-custom-sudo`: Allows all `sudo` users to escalate

To use this role set Tag to `hardening`

> 🔒 This is important to prevent privilege escalation by unauthorized users.

---

### 3. `bind9`: Install and Configure a DNS Server

This role:

* Installs BIND9 and its utilities
* Deploys a custom `named.conf.options` config from Jinja2 template
* Ensures BIND9 is running and enabled on boot

System DNS is also updated to use custom DNS servers (e.g., Shecan).

To use this role set Tag to `bind9`

---

### 4. `gitlab`: GitLab + Registry + Runner (Docker Stack)

This role:

* Installs Docker and Docker Compose
* Adds the Ansible user to the Docker group
* Pulls required images:

  * GitLab CE
  * GitLab Registry
  * GitLab Runner
* Runs each service in its own container on a Docker network (`gitlab-network`)
* Configures GitLab via environment variables
* Extracts and stores the GitLab root password
* Registers the GitLab Runner **automatically**

To use this role set Tag to `GitLab,Registry,Runner`

**Volumes** are mounted to persist GitLab, Registry, and Runner configuration.

---

## User Management and Hardening

### User Management

* **User Creation**: Ansible creates a specified user (default: `ubuntu`).
* **Shell & Home**: Sets `/bin/bash` as default shell.
* **SSH Access**: Adds a public SSH key from the controller to `~/.ssh/authorized_keys`.
* **Permissions**:

  * `.ssh/` directory is set to `0700`
  * SSH key file is managed securely

**Benefit**: Prevents use of passwords over SSH, which is more secure and automation-friendly.

---

### Security Hardening

* **Sudo Enforcement**:

  * Installs `sudo` package
  * Grants the user access via the `sudo` group
* **Disable `su`**:

  * Updates PAM configuration to restrict `su` to members of the `sudo` group only
* **Sudoers Config**:

  * Custom sudoers file allows all `sudo` group members to run any command as root with password
  * Config is placed in `/etc/sudoers.d` for modular control

**Goal**: Ensure only explicitly authorized users can perform administrative tasks, and prevent privilege escalation through `su`.

---

## 🛠 Example Variables (group\_vars or `extra_vars`)

```yaml
user_name: ubuntu
public_key_file: ~/.ssh/id_rsa.pub
dns:
  - 178.22.122.100
  - 185.51.200.2
gitlab_runner_token: YOUR_REGISTRATION_TOKEN_HERE
docker_token: YOUR_DOCKER_LOGIN_TOKEN_HERE
```
