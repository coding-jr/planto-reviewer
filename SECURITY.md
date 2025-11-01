# Security Guide

This document outlines security best practices for the Code Quality Tracker application.

## 🔒 Credential Management

### Never Commit Secrets to Git

The following files are git-ignored to prevent credential exposure:

- `settings.json` - Contains AWS credentials and API keys
- `backend/.env` - Environment-specific configuration
- `.env.local`, `.env.production` - Environment files

### Use Example Files as Templates

Always use the example files and copy them locally:

```bash
# For backend environment variables
cp backend/.env.example backend/.env

# For settings.json configuration
cp settings.example.json settings.json
```

Then edit these files with your actual credentials. **Never commit these files to git.**

## 🔑 Supported Environment Variables

The application supports the following environment variables (in priority order):

### Database Configuration
- `DATABASE_URL` - MySQL connection string

### API Configuration
- `PORT` - Server port (default: 3000)
- `API_KEY` - API authentication key
- `ENV` - Environment (development, production)

### AI Provider Configuration
- `AI_PROVIDER` - Provider selection (bedrock, openai, anthropic, google)
- `AI_API_KEY` - API key for non-Bedrock providers
- `AI_MODEL` - Model name

### AWS Bedrock Configuration

#### Authentication (Choose one method):

**Method 1: Bearer Token (Recommended)**
- `AWS_BEARER_TOKEN_BEDROCK` - Bedrock API Key (preferred name, matches VSCode settings)
- `AWS_BEARER_TOKEN` - Alternative name for bearer token

**Method 2: IAM Credentials**
- `AWS_ACCESS_KEY_ID` - AWS access key ID
- `AWS_SECRET_ACCESS_KEY` - AWS secret access key

#### Model Configuration:
- `AWS_REGION` - AWS region (e.g., ap-south-1, us-east-1)
- `BEDROCK_MODEL` - Model ID (e.g., global.anthropic.claude-sonnet-4-5-20250929-v1:0)
- `BEDROCK_MODEL_ARN` - Full ARN for inference profile (optional)
- `BEDROCK_MAX_TOKENS` - Maximum tokens per request (default: 16192)
- `BEDROCK_TEMPERATURE` - Temperature for AI responses (default: 0.3)

### Worker Configuration
- `POLLING_INTERVAL` - GitHub polling interval in seconds (default: 30)

## 🎯 Configuration Priority

The application loads configuration in the following priority order:

1. **Environment Variables** (highest priority)
2. **settings.json** (fallback)
3. **Default values** (lowest priority)

This allows you to:
- Use `settings.json` for non-sensitive defaults
- Override sensitive values with environment variables
- Keep credentials out of version control

## 🔐 How to Get AWS Bedrock API Key

### Option 1: Bedrock API Key (Recommended)

1. Sign in to **AWS Console**
2. Navigate to **Amazon Bedrock** service
3. Go to **API Keys** section
4. Click **"Create API Key"**
5. Copy the bearer token (starts with `ABSK...`)
6. Set it as `AWS_BEARER_TOKEN_BEDROCK` in your environment

**Advantages:**
- Simpler setup
- No IAM user management needed
- Scoped to Bedrock only

### Option 2: IAM Credentials

1. Sign in to **AWS Console**
2. Navigate to **IAM** → **Users**
3. Create a new user or select existing user
4. Attach policy: `AmazonBedrockFullAccess`
5. Create access keys
6. Set `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY`

**Advantages:**
- More control with IAM policies
- Can be used across multiple AWS services

## 🛡️ Security Best Practices

### ✅ DO:

1. **Use environment variables** for all sensitive data
2. **Rotate credentials regularly** (every 90 days)
3. **Use strong, random API keys** for your API
4. **Keep `.env` and `settings.json` in `.gitignore`**
5. **Use different credentials** for development and production
6. **Review access logs** regularly
7. **Enable MFA** on AWS accounts
8. **Use least privilege principle** for IAM policies
9. **Monitor AWS Bedrock usage** for anomalies
10. **Keep dependencies updated** (`go mod tidy`, `npm update`)

### ❌ DON'T:

1. **Never commit credentials** to version control
2. **Never share credentials** in chat, email, or documentation
3. **Never use default or weak passwords**
4. **Never store credentials** in code comments
5. **Never expose API keys** in client-side code
6. **Never use production credentials** in development
7. **Never log sensitive data** (API keys, tokens, passwords)
8. **Never share `.env` files** or `settings.json`

## 🚨 If Credentials Are Exposed

If you accidentally commit credentials to git:

### 1. Immediately Rotate Credentials

- Generate new AWS Bedrock API keys
- Update your local `.env` or `settings.json`
- Delete old credentials from AWS Console

### 2. Remove from Git History

```bash
# Install git-filter-repo (recommended)
# On macOS:
brew install git-filter-repo

# On Linux:
pip install git-filter-repo

# Remove sensitive files from git history
git filter-repo --path settings.json --invert-paths
git filter-repo --path backend/.env --invert-paths

# Force push (⚠️ WARNING: This rewrites history)
git push origin --force --all
```

**Alternative method using BFG Repo Cleaner:**

```bash
# Download BFG from https://rtyley.github.io/bfg-repo-cleaner/

# Remove specific text (e.g., API key)
java -jar bfg.jar --replace-text secrets.txt

# Clean up
git reflog expire --expire=now --all
git gc --prune=now --aggressive

# Force push
git push origin --force --all
```

### 3. Notify Your Team

If this is a shared repository:
- Inform all team members
- Ensure they pull the cleaned history
- Verify no one has old credentials cached

### 4. Monitor for Misuse

- Check AWS CloudTrail logs
- Monitor for unusual API activity
- Review billing for unexpected charges

## 📋 Security Checklist

Before deploying to production:

- [ ] All credentials are in environment variables or git-ignored files
- [ ] `.gitignore` includes `settings.json`, `backend/.env`, and `.env*`
- [ ] No secrets in git history (`git log -p | grep -i "secret"`)
- [ ] Strong API keys generated for `API_KEY`
- [ ] AWS Bedrock credentials are valid and scoped correctly
- [ ] Database uses strong passwords
- [ ] HTTPS enabled for production API
- [ ] CORS configured correctly
- [ ] Rate limiting enabled
- [ ] Logging configured (without sensitive data)
- [ ] Backup strategy in place
- [ ] Monitoring and alerting configured
- [ ] Dependencies audited (`npm audit`, `go mod verify`)

## 🔍 Auditing Your Configuration

Run these commands to check for exposed credentials:

```bash
# Check git history for potential secrets
git log -p | grep -i "bearer"
git log -p | grep -i "ABSK"
git log -p | grep -i "secret"

# Check current files
grep -r "ABSK" . --exclude-dir=node_modules --exclude-dir=.git
grep -r "aws_secret" . --exclude-dir=node_modules --exclude-dir=.git

# Verify .gitignore is working
git check-ignore settings.json backend/.env
# Should return the file paths if properly ignored
```

## 📞 Reporting Security Issues

If you discover a security vulnerability:

1. **Do NOT** create a public GitHub issue
2. Email the maintainers directly
3. Include detailed steps to reproduce
4. Allow time for a fix before public disclosure

## 📚 Additional Resources

- [AWS Bedrock Security Best Practices](https://docs.aws.amazon.com/bedrock/latest/userguide/security.html)
- [OWASP Secrets Management Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html)
- [GitHub: Removing sensitive data](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/removing-sensitive-data-from-a-repository)
