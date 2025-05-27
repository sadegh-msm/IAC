#!/bin/bash

if [ -f /etc/gitlab-runner/config.toml ]; then
    echo "Runner already registered"
    exit 0
fi

gitlab-runner register --non-interactive \
  --url "http://{{ ansible_host }}" \
  --registration-token "{{ gitlab_runner_registration_token }}" \
  --executor "docker" \
  --docker-image "alpine:latest" \
  --description "docker-runner" \
  --tag-list "docker,linux" \
  --run-untagged="true" \
  --locked="false"

