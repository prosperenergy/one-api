#!/usr/bin/env bash
set -euo pipefail

# Local helper to build image and deploy to Cloud Run using gcloud CLI
# Usage:
#   GCP_PROJECT=your-project CLOUD_RUN_REGION=us-central1 ONEAPI_JWT_SECRET=... ./scripts/deploy-cloud-run.sh

if ! command -v gcloud >/dev/null 2>&1; then
  echo "gcloud not installed. Install Google Cloud SDK: https://cloud.google.com/sdk/docs/install" >&2
  exit 1
fi

PROJECT=${GCP_PROJECT:-}
if [ -z "$PROJECT" ]; then
  echo "Set GCP_PROJECT env var or pass it: GCP_PROJECT=your-project ./scripts/deploy-cloud-run.sh" >&2
  exit 1
fi

REGION=${CLOUD_RUN_REGION:-us-central1}
IMAGE=gcr.io/$PROJECT/one-api:local-$(date +%s)

# Build
docker build -t "$IMAGE" .

# Ensure gcloud Docker auth configured
gcloud auth configure-docker --quiet

# Push
docker push "$IMAGE"

# Deploy
gcloud run deploy one-api \
  --image "$IMAGE" \
  --region "$REGION" \
  --platform managed \
  --allow-unauthenticated \
  --set-env-vars NODE_ENV=production,ONEAPI_JWT_SECRET=${ONEAPI_JWT_SECRET:-} \
  --project "$PROJECT"

echo "Deployed $IMAGE to Cloud Run in project $PROJECT (region $REGION)"
