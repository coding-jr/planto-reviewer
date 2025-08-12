import { BedrockRuntimeClient, InvokeModelCommand } from '@aws-sdk/client-bedrock-runtime';
import { BaseChatModel } from 'langchain/chat_models/base';
import type { BaseChatMessage, ChatGeneration, ChatResult, LLMResult } from 'langchain/schema';
import { AIChatMessage, HumanChatMessage, SystemChatMessage } from 'langchain/schema';
import { CallbackManagerForLLMRun } from 'langchain/callbacks';

interface ClaudeBedrockConfig {
  region: string;
  modelId: string;
  accessKeyId?: string;
  secretAccessKey?: string;
  maxTokens?: number;
  temperature?: number;
}

interface ClaudeMessage {
  role: 'user' | 'assistant';
  content: string;
}

interface ClaudeRequest {
  anthropic_version: string;
  max_tokens: number;
  temperature?: number;
  system?: string;
  messages: ClaudeMessage[];
}

interface ClaudeResponse {
  id: string;
  type: string;
  role: string;
  content: Array<{
    type: string;
    text: string;
  }>;
  model: string;
  stop_reason: string;
  stop_sequence: null;
  usage: {
    input_tokens: number;
    output_tokens: number;
  };
}

export class ClaudeBedrockService extends BaseChatModel {
  private client: BedrockRuntimeClient;
  private config: ClaudeBedrockConfig;

  _combineLLMOutput?(...llmOutputs: LLMResult['llmOutput'][]): LLMResult['llmOutput'] {
    return llmOutputs.reduce((acc, output) => {
      if (output?.tokenUsage) {
        acc = acc || { tokenUsage: { promptTokens: 0, completionTokens: 0, totalTokens: 0 } };
        acc.tokenUsage!.promptTokens += output.tokenUsage.promptTokens || 0;
        acc.tokenUsage!.completionTokens += output.tokenUsage.completionTokens || 0;
        acc.tokenUsage!.totalTokens += output.tokenUsage.totalTokens || 0;
      }
      return acc;
    }, {} as LLMResult['llmOutput']);
  }

  constructor(config: ClaudeBedrockConfig) {
    super({});
    this.config = {
      maxTokens: 4096,
      temperature: 0,
      ...config
    };
    
    this.client = new BedrockRuntimeClient({
      region: this.config.region,
      credentials: this.config.accessKeyId && this.config.secretAccessKey ? {
        accessKeyId: this.config.accessKeyId,
        secretAccessKey: this.config.secretAccessKey
      } : undefined
    });
  }

  _llmType(): string {
    return 'claude-bedrock';
  }

  async _generate(
    messages: BaseChatMessage[],
    options?: { stop?: string[] },
    runManager?: CallbackManagerForLLMRun
  ): Promise<ChatResult> {
    const claudeRequest = this.convertMessagesToClaudeFormat(messages);
    
    try {
      const command = new InvokeModelCommand({
        modelId: this.config.modelId,
        contentType: 'application/json',
        accept: 'application/json',
        body: JSON.stringify(claudeRequest)
      });

      const response = await this.client.send(command);
      const responseBody = JSON.parse(new TextDecoder().decode(response.body));
      
      return this.handleClaudeResponse(responseBody);
    } catch (error) {
      console.error('Error calling Claude Bedrock:', error);
      throw error;
    }
  }

  private convertMessagesToClaudeFormat(messages: BaseChatMessage[]): ClaudeRequest {
    let systemMessage = '';
    const claudeMessages: ClaudeMessage[] = [];

    for (const message of messages) {
      if (message._getType() === 'system') {
        // Claude expects system message to be separate from the messages array
        systemMessage = (message as SystemChatMessage).text;
      } else if (message._getType() === 'human') {
        claudeMessages.push({
          role: 'user',
          content: (message as HumanChatMessage).text
        });
      } else if (message._getType() === 'ai') {
        claudeMessages.push({
          role: 'assistant',
          content: (message as AIChatMessage).text
        });
      }
    }

    const request: ClaudeRequest = {
      anthropic_version: '2023-06-01',
      max_tokens: this.config.maxTokens!,
      temperature: this.config.temperature,
      messages: claudeMessages
    };

    if (systemMessage) {
      request.system = systemMessage;
    }

    return request;
  }

  private handleClaudeResponse(response: ClaudeResponse): ChatResult {
    const text = response.content[0]?.text || '';
    
    const generation: ChatGeneration = {
      text,
      message: new AIChatMessage(text)
    };

    return {
      generations: [generation],
      llmOutput: {
        tokenUsage: {
          promptTokens: response.usage?.input_tokens || 0,
          completionTokens: response.usage?.output_tokens || 0,
          totalTokens: (response.usage?.input_tokens || 0) + (response.usage?.output_tokens || 0)
        }
      }
    };
  }
}
