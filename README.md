# kk - kubectl extension

## Overview

`kk` is a command-line utility designed to extend the functionality of `kubectl`. It aims to streamline common Kubernetes workflows and improve user experience, particularly around managing Kubernetes contexts.

While `kubectl` is powerful, managing numerous contexts or dealing with long, auto-generated context names (common with cloud providers like GKE, EKS, AKS) can become cumbersome. `kk` provides helpful additions to address these points, making context switching faster and more intuitive.

## Features

* **Interactive Context Switching**: Easily switch between your Kubernetes contexts using an interactive fuzzy finder interface.
* **Context Aliases**: Create and use custom, shorter aliases for your lengthy or complex context names.
* **Fuzzy Finder**: Quickly find the context or alias you're looking for, even with many entries.

**Example Usage:**

### Switch context (by alias)

Run `kk context` to launch the interactive context switcher:

```bash
kk context
```

This command provides an enhanced alternative to `kubectl config use-context`.

### Resolve context alias

Run `kk context resolve "your-alias"` to resolve alias into real context name.

```bash
kk context resolve "your-alias"
```

This command prints context aliased by name "your-alias" to standard output.

### Display resource quota

Run `kk quota [namespace]` to display resource quota for a namespace in a human-readable format.

```bash
kk quota
# or
kk quota my-namespace
```

This command fetches and displays resource quota information for the specified namespace (or current namespace if none specified). It shows resource usage including CPU requests/limits, memory requests/limits, and storage requests. The output includes:
- NAME: The resource name (e.g., requests.cpu, limits.memory)
- USED: Currently used amount of the resource
- HARD: Maximum allowed amount of the resource (quota limit)
- USAGE: Percentage of used resources relative to the quota limit

## Installation

**Prerequisites:**

* Go (Golang) must be installed on your system.

**Command:**

Install the latest version using `go install`:

```bash
go install github.com/piotrszyma/kk@latest
```

Ensure your Go bin directory (usually `$HOME/go/bin`) is in your system's `$PATH`.

## Configuration

`kk` uses a configuration file primarily to define custom aliases for your Kubernetes context names.

* **Configuration File Location**: Create or edit the configuration file at:
    ```
    ~/.config/kk/config.yaml
    ```
    If the file or the `.config/kk` directory does not exist when `kk` is run, it will be created automatically with default values (an empty configuration).

* **Configuration Structure**:

    The configuration file uses YAML format. To define aliases, add entries under `context.aliases`. Each entry requires a `name` (the actual Kubernetes context name) and an `alias` (the custom name you want to use).

* **Example `~/.config/kk/config.yaml`**:

    ```yaml
    # ~/.config/kk/config.yaml
    context:
      aliases:
        # Alias "prod-main" for the context "gke_my-project-123_us-central1-a_my-main-cluster"
        - name: gke_my-project-123_us-central1-a_my-main-cluster
          alias: prod-main

        # Alias "dev-user1" for the context "arn:aws:eks:eu-west-1:123456789012:cluster/development-cluster-user1"
        - name: arn:aws:eks:eu-west-1:123456789012:cluster/development-cluster-user1
          alias: dev-user1

        # You can have multiple aliases for the same context name if needed
        - name: gke_my-project-123_us-central1-a_my-main-cluster
          alias: main-gke-prod
    ```

When you run `kk context`, both the original context names and your defined aliases will appear in the fuzzy finder, allowing you to select either. Selecting an alias will switch the Kubernetes context to the corresponding original context name.
