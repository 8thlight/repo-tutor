# 0010. Agent driver

Status: accepted

## Context

The tutor needs a model. Linking an API SDK means key handling. Relying on one vendor CLI's headless mode is a risk if that mode changes.

## Decision

The binary runs the tutor through an agent driver. A driver starts one headless turn in a named session, streams the reply, and returns it. Two drivers ship first: `claude -p` and `codex exec`, with resume for follow-up turns. The learner picks one. The core never knows which.

The binary does not call a model API. Quiz outcomes come back as structured output the binary parses and stores.
