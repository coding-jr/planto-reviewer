import { BedrockRuntimeClient, InvokeModelCommand } from '@aws-sdk/client-bedrock-runtime';

async function main() {
  const client = new BedrockRuntimeClient({
    region: process.env.AWS_REGION,
    credentials: {
      accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
      secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!
    },
  });

  const modelId = process.env.BEDROCK_MODEL_ID || 'anthropic.claude-3-sonnet-20240229-v1:0';

  const prompt = [
    {
      role: 'user',
      content: 'Who won the FIFA World Cup in 2022?'
    }
  ];

  const request = {
    anthropic_version: 'bedrock-2023-05-31',
    max_tokens: 256,
    temperature: 0.7,
    system: `You are a helpful and factual assistant.\nReturn only the final answer, no preamble.`,
    messages: prompt,
  };

  const input: any = {
    modelId,
    contentType: 'application/json',
    accept: 'application/json',
    body: JSON.stringify(request),
  };

  const command = new InvokeModelCommand(input);

  try {
    const response = await client.send(command);
    const responseBody = JSON.parse(new TextDecoder().decode(response.body));
    console.log('Claude output:', responseBody.content?.[0]?.text || responseBody);
  } catch (err) {
    console.error('Claude test error:', err);
    process.exitCode = 1;
  }
}

main();

