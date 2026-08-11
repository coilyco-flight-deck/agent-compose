#!/bin/sh
# Run the board through Inspect against Agent Proxy. Unscored on purpose: the
# scorer is a human, and Inspect calls the repetition an epoch.
set -e
samples=${EVAL_SAMPLES:-samples.yaml}
prompts=${EVAL_PROMPTS:-.evalkit/prompts}
logs=${EVAL_LOGS:-.evalkit/logs}
epochs=${EVAL_EPOCHS:-5}

AGENTPROXY_BASE_URL=${AGENT_PROXY_BASE:-http://ser8:8080/v1}
AGENTPROXY_API_KEY=${AGENTPROXY_API_KEY:-unused}
export AGENTPROXY_BASE_URL AGENTPROXY_API_KEY

model=${AGENT_PROXY_MODEL:-evaluation/deepseek-v4-flash}
uv run inspect eval evalkit/task.py \
  --model "openai-api/agentproxy/$model" \
  --epochs "$epochs" \
  --no-score \
  --log-dir "$logs" \
  -T samples="$samples" \
  -T prompts="$prompts" \
  "$@"
