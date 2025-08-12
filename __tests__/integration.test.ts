import { jest, describe, it, expect, beforeEach } from '@jest/globals';
import { ClaudeBedrockService } from '../src/services/claudeBedrockService';
import { BedrockRuntimeClient, InvokeModelCommand } from '@aws-sdk/client-bedrock-runtime';
import { SystemChatMessage, HumanChatMessage } from 'langchain/schema';

// Mock AWS SDK
jest.mock('@aws-sdk/client-bedrock-runtime');

describe('Claude Message Format Verification', () => {
  let mockClient: any;
  let claudeService: ClaudeBedrockService;

  beforeEach(() => {
    mockClient = { 
      send: jest.fn().mockImplementation((command) => {
        // Store the command for test verification
        mockClient.lastCommand = command;
        return Promise.resolve({
          body: new TextEncoder().encode(JSON.stringify({
            content: [{ text: 'Mock response' }],
            usage: { input_tokens: 50, output_tokens: 20 }
          }))
        });
      })
    };
    (BedrockRuntimeClient as jest.Mock).mockImplementation(() => mockClient);

    claudeService = new ClaudeBedrockService({
      region: 'us-east-1',
      modelId: 'anthropic.claude-3-sonnet-20240229-v1:0',
      accessKeyId: 'test-key',
      secretAccessKey: 'test-secret',
      temperature: 0.7,
      maxTokens: 4096
    });
  });

  it('should generate proper Claude API message format for code review', async () => {
    // Mock Claude's response
    const mockResponse = {
      body: new TextEncoder().encode(JSON.stringify({
        id: 'msg_test',
        type: 'message',
        role: 'assistant',
        content: [{
          type: 'text',
          text: 'Code review completed successfully.'
        }],
        model: 'claude-3-sonnet-20240229',
        stop_reason: 'end_turn',
        stop_sequence: null,
        usage: {
          input_tokens: 100,
          output_tokens: 20
        }
      }))
    };

    (mockClient.send as any).mockResolvedValue(mockResponse);

    const messages = [
      new SystemChatMessage("Act as a code reviewer."),
      new HumanChatMessage("Please review this code: function test() { return true; }")
    ];

    const result = await claudeService._generate(messages as any);

    // Verify the Claude API was called with correct format
    expect(mockClient.send).toHaveBeenCalledTimes(1);
    expect(mockClient.send).toHaveBeenCalledWith(
      expect.any(InvokeModelCommand)
    );

    // Verify the response is properly formatted
    expect(result.generations[0].text).toBe('Code review completed successfully.');
    expect(result.llmOutput?.tokenUsage).toEqual({
      promptTokens: 100,
      completionTokens: 20,
      totalTokens: 120
    });
  });

});
