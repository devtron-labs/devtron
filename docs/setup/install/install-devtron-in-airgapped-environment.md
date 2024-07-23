# Devtron Full Mode Installation in an Airgapped Environment

## Pre-requisites

1. Install `podman` or `docker` and `yq` on the VM from where you're executing the installation commands.
2. Clone the Devtron Helm chart:

    ```bash
    git clone https://github.com/devtron-labs/devtron.git
    cd devtron
    ```


4. Set the value of `TARGET_REGISTRY`, `TARGET_REGISTRY_USERNAME`, and `TARGET_REGISTRY_TOKEN`. This registry should be accessible from the VM where you are running the cloning script and the K8s cluster where you’re installing Devtron.

## Docker Instructions

### For Linux/amd64

1. Set the environment variables:

    ```bash
    # Set the source registry URL
    export SOURCE_REGISTRY="quay.io/devtron"

    # Set the target registry URL, username, and token/password
    export TARGET_REGISTRY_URL=""
    export TARGET_REGISTRY_USERNAME=""
    export TARGET_REGISTRY_TOKEN=""

    # Set the source and target image file names with default values if not already set
    SOURCE_IMAGES_LIST="${SOURCE_IMAGES_LIST:=devtron-images.txt.source}"
    TARGET_IMAGES_LIST="${TARGET_IMAGES_LIST:=devtron-images.txt.target}"
    ```

2. Log in to the target Docker registry:

    ```bash
    docker login -u $TARGET_REGISTRY_USERNAME -p $TARGET_REGISTRY_TOKEN $TARGET_REGISTRY_URL
    ```

3. Clone the images:

    ```bash
    while IFS= read -r source_image; do
      # Check if the source image belongs to the quay.io/devtron registry
      if [[ "$source_image" == quay.io/devtron/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/quay.io\/devtron/$TARGET_REGISTRY_URL}"

      # Check if the source image belongs to the quay.io/argoproj registry
      elif [[ "$source_image" == quay.io/argoproj/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/quay.io\/argoproj/$TARGET_REGISTRY_URL}"

      # Check if the source image belongs to the public.ecr.aws/docker/library registry
      elif [[ "$source_image" == public.ecr.aws/docker/library/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/public.ecr.aws\/docker\/library/$TARGET_REGISTRY_URL}"
      fi

      # Pull the image from the source registry
      docker pull --platform linux/amd64 $source_image

      # Tag the image with the new target registry name
      docker tag $source_image $target_image

      # Push the image to the target registry
      docker push $target_image

      # Output the updated image name
      echo "Updated image: $target_image"

      # Append the new image name to the target images file
      echo "$target_image" >> "$TARGET_IMAGES_LIST"

    done < "$SOURCE_IMAGES_LIST"
    ```

### For Linux/arm64

1. Set the environment variables:

    ```bash
    # Set the source registry URL
    export SOURCE_REGISTRY="quay.io/devtron"

    # Set the target registry URL, username, and token/password
    export TARGET_REGISTRY_URL=""
    export TARGET_REGISTRY_USERNAME=""
    export TARGET_REGISTRY_TOKEN=""

    # Set the source and target image file names with default values if not already set
    SOURCE_IMAGES_LIST="${SOURCE_IMAGES_LIST:=devtron-images.txt.source}"
    TARGET_IMAGES_LIST="${TARGET_IMAGES_LIST:=devtron-images.txt.target}"
    ```

2. Log in to the target Docker registry:

    ```bash
    docker login -u $TARGET_REGISTRY_USERNAME -p $TARGET_REGISTRY_TOKEN $TARGET_REGISTRY_URL
    ```

3. Clone the images:

    ```bash
    while IFS= read -r source_image; do
      # Check if the source image belongs to the quay.io/devtron registry
      if [[ "$source_image" == quay.io/devtron/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/quay.io\/devtron/$TARGET_REGISTRY_URL}"

      # Check if the source image belongs to the quay.io/argoproj registry
      elif [[ "$source_image" == quay.io/argoproj/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/quay.io\/argoproj/$TARGET_REGISTRY_URL}"

      # Check if the source image belongs to the public.ecr.aws/docker/library registry
      elif [[ "$source_image" == public.ecr.aws/docker/library/* ]]; then
        # Replace the source registry with the target registry in the image name
        target_image="${source_image/public.ecr.aws\/docker\/library/$TARGET_REGISTRY_URL}"
      fi

      # Pull the image from the source registry
      docker pull --platform linux/arm64 $source_image

      # Tag the image with the new target registry name
      docker tag $source_image $target_image

      # Push the image to the target registry
      docker push $target_image

      # Output the updated image name
      echo "Updated image: $target_image"

      # Append the new image name to the target images file
      echo "$target_image" >> "$TARGET_IMAGES_LIST"

    done < "$SOURCE_IMAGES_LIST"
    ```

## Podman Instructions

### For Multi-arch

1. Set the environment variables:

    ```bash
    export SOURCE_REGISTRY="quay.io/devtron"
    export SOURCE_REGISTRY_TOKEN=#Enter token provided by Devtron team
    export TARGET_REGISTRY=#Enter target registry url 
    export TARGET_REGISTRY_USERNAME=#Enter target registry username 
    export TARGET_REGISTRY_TOKEN=#Enter target registry token/password
    ```

2. Log in to the target Podman registry:

    ```bash
    podman login -u $TARGET_REGISTRY_USERNAME -p $TARGET_REGISTRY_TOKEN $TARGET_REGISTRY
    ```

3. Clone the images:

    ```bash
    SOURCE_REGISTRY="quay.io/devtron"
    TARGET_REGISTRY=${TARGET_REGISTRY}
    SOURCE_IMAGES_FILE_NAME="${SOURCE_IMAGES_FILE_NAME:=devtron-images.txt.source}"
    TARGET_IMAGES_FILE_NAME="${TARGET_IMAGES_FILE_NAME:=devtron-images.txt.target}"

    cp $SOURCE_IMAGES_FILE_NAME $TARGET_IMAGES_FILE_NAME
    while read source_image; do
      if [[ "$source_image" == *"workflow-controller:"* || "$source_image" == *"argoexec:"* || "$source_image" == *"argocd:"* ]]
      then
        SOURCE_REGISTRY="quay.io/argoproj"
        sed -i "s|${SOURCE_REGISTRY}|${TARGET_REGISTRY}|g" $TARGET_IMAGES_FILE_NAME
      elif [[ "$source_image" == *"redis:"* ]]
      then
        SOURCE_REGISTRY="public.ecr.aws/docker/library"
        sed -i "s|${SOURCE_REGISTRY}|${TARGET_REGISTRY}|g" $TARGET_IMAGES_FILE_NAME
      else
        SOURCE_REGISTRY="quay.io/devtron"
        sed -i "s|${SOURCE_REGISTRY}|${TARGET_REGISTRY}|g" $TARGET_IMAGES_FILE_NAME
      fi
    done <$SOURCE_IMAGES_FILE_NAME
    echo "Target Images file finalized"

    while read -r -u 3 source_image && read -r -u 4 target_image ; do
      echo "Pushing $source_image $target_image"
      podman manifest create $source_image
      podman manifest add $source_image $source_image --all
      podman manifest push $source_image $target_image --all
    done 3<"$SOURCE_IMAGES_FILE_NAME" 4<"$TARGET_IMAGES_FILE_NAME"
    ```

## Final Step

Change your registry name in `values.yaml` at line number 10:

```bash
yq e '.global.containerRegistry = "<your-registry-name>"' -i /charts/devtron/values.yaml
