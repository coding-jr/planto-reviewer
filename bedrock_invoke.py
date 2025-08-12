#!/usr/bin/env python3
import json
import os
import sys
import botocore
import boto3


def eprint(*args, **kwargs):
    print(*args, file=sys.stderr, **kwargs)


def main():
    arn = os.getenv("BEDROCK_INFERENCE_PROFILE_ARN") or os.getenv("INFERENCE_PROFILE_ARN")
    prompt = os.getenv("PROMPT", "Say hello in one short sentence.")
    region = boto3.session.Session().region_name

    if not arn:
        eprint("ERROR: BEDROCK_INFERENCE_PROFILE_ARN env var is not set.")
        eprint("Set it, then re-run. Example:")
        eprint("  export BEDROCK_INFERENCE_PROFILE_ARN=arn:aws:bedrock:us-east-1:123456789012:inference-profile/my-profile")
        sys.exit(2)

    if not region:
        eprint("WARNING: No AWS region configured in your environment. Falling back to us-east-1.")
        region = "us-east-1"

    session = boto3.session.Session(region_name=region)
    bedrock_rt = session.client("bedrock-runtime")
    bedrock_ctrl = session.client("bedrock")

    # Optional: verify the inference profile exists and we can read it
    try:
        # get_inference_profile may not be available in all SDK versions; handle gracefully
        if hasattr(bedrock_ctrl, "get_inference_profile"):
            prof = bedrock_ctrl.get_inference_profile(inferenceProfileIdentifier=arn)  # type: ignore[arg-type]
            print("Verified inference profile:")
            print(json.dumps(prof.get("inferenceProfile", prof), indent=2, default=str))
        else:
            print("SDK does not expose get_inference_profile; skipping profile pre-check.")
    except botocore.exceptions.ClientError as e:
        eprint("WARN: Could not describe the inference profile (may be a permissions or SDK-version issue):")
        eprint(str(e))

    # Try converse API first (modelId can be an ARN or alias)
    try:
        print("\nAttempting bedrock-runtime.converse via inference profile ARN ...")
        resp = bedrock_rt.converse(
            modelId=arn,
            messages=[{"role": "user", "content": [{"text": prompt}]}],
            inferenceConfig={"maxTokens": 128},
        )
        output = resp.get("output", {})
        if not output:
            print(json.dumps(resp, indent=2, default=str))
        else:
            # Extract text blocks if present
            parts = output.get("message", {}).get("content", [])
            texts = []
            for p in parts:
                if "text" in p:
                    texts.append(p["text"])
            print("Converse response:")
            print("\n".join(texts) if texts else json.dumps(resp, indent=2, default=str))
        return
    except botocore.exceptions.ClientError as e:
        eprint("Converse attempt failed:")
        eprint(str(e))
    except Exception as e:
        eprint("Converse attempt error:")
        eprint(repr(e))

    # Fallback: invoke_model with an Anthropic-compatible JSON body.
    # This will work if the routed model via the profile supports Claude Messages API.
    try:
        print("\nAttempting bedrock-runtime.invoke_model with Anthropic-style payload ...")
        body = {
            "anthropic_version": "bedrock-2023-05-31",
            "max_tokens": 128,
            "messages": [
                {"role": "user", "content": [{"type": "text", "text": prompt}]}
            ],
        }
        resp = bedrock_rt.invoke_model(
            modelId=arn,
            contentType="application/json",
            accept="application/json",
            body=json.dumps(body).encode("utf-8"),
        )
        payload = resp.get("body")
        if hasattr(payload, "read"):
            data = payload.read()
        else:
            data = payload
        print("InvokeModel response:")
        print(data.decode("utf-8") if isinstance(data, (bytes, bytearray)) else str(data))
        return
    except botocore.exceptions.ClientError as e:
        eprint("InvokeModel attempt failed:")
        eprint(str(e))
    except Exception as e:
        eprint("InvokeModel attempt error:")
        eprint(repr(e))

    eprint("\nAll attempts failed. Diagnostics:")
    try:
        sts = session.client("sts")
        ident = sts.get_caller_identity()
        print(json.dumps({"caller": ident, "region": region, "inference_profile_arn": arn}, indent=2))
    except Exception:
        pass
    eprint("Checklist: \n- Ensure the ARN is correct and in this region.\n- Confirm IAM permissions: bedrock:InvokeModel and bedrock:Converse on the target inference profile or routed models.\n- Verify that the routed model for the profile supports the chosen API (Converse or Anthropic Messages).\n- Confirm VPC endpoints or network egress allow Bedrock access if using private networking.")
    sys.exit(1)


if __name__ == "__main__":
    main()

