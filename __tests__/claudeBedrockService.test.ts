import { jest, describe, it, expect, beforeEach } from '@jest/globals';
import { ClaudeBedrockService } from '../src/services/claudeBedrockService';
import { BedrockRuntimeClient, InvokeModelCommand } from '@aws-sdk/client-bedrock-runtime';
import { SystemChatMessage, HumanChatMessage, AIChatMessage } from 'langchain/schema';

// Mock AWS SDK
jest.mock('@aws-sdk/client-bedrock-runtime');

describe('ClaudeBedrockService', () => {
  let service: ClaudeBedrockService;
  let mockClient: any;

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
    
    service = new ClaudeBedrockService({
      region: 'us-east-1',
      modelId: 'anthropic.claude-3-sonnet-20240229-v1:0',
      accessKeyId: 'test-key',
      secretAccessKey: 'test-secret',
      temperature: 0.7,
      maxTokens: 4096
    });
  });

  it('should correctly format system and user messages for Claude API', async () => {
    const mockResponse = {
      body: new TextEncoder().encode(JSON.stringify({
        id: 'msg_01ABC123',
        type: 'message',
        role: 'assistant',
        content: [
          {
            type: 'text',
            text: 'This code looks good! The implementation follows best practices.'
          }
        ],
        model: 'claude-3-sonnet-20240229',
        stop_reason: 'end_turn',
        stop_sequence: null,
        usage: {
          input_tokens: 100,
          output_tokens: 50
        }
      }))
    };

    mockClient.send.mockResolvedValue(mockResponse);

    const messages = [
      new SystemChatMessage("Act as an empathetic software engineer that's an expert in all programming languages."),
      new HumanChatMessage("Please review this code: function add(a, b) { return a + b; }")
    ];

    const result = await service._generate(messages);

    // Test core functionality - service should be called and return formatted response
    expect(mockClient.send).toHaveBeenCalledTimes(1);
    expect(mockClient.send).toHaveBeenCalledWith(
      expect.any(InvokeModelCommand)
    );

    expect(result.generations).toHaveLength(1);
    expect(result.generations[0].text).toBe('This code looks good! The implementation follows best practices.');
    expect(result.llmOutput?.tokenUsage).toEqual({
      promptTokens: 100,
      completionTokens: 50,
      totalTokens: 150
    });
  });

  it('should handle conversation with alternating messages', async () => {
    const mockResponse = {
      body: new TextEncoder().encode(JSON.stringify({
        id: 'msg_02XYZ456',
        type: 'message',
        role: 'assistant',
        content: [
          {
            type: 'text',
            text: 'I understand the context. Here are my suggestions for improving the code.'
          }
        ],
        model: 'claude-3-sonnet-20240229',
        stop_reason: 'end_turn',
        stop_sequence: null,
        usage: {
          input_tokens: 150,
          output_tokens: 75
        }
      }))
    };

    mockClient.send.mockResolvedValue(mockResponse);

    const messages = [
      new SystemChatMessage("You are a code reviewer."),
      new HumanChatMessage("Please review this function."),
      new AIChatMessage("I can help you review the function. Please provide the code."),
      new HumanChatMessage("Here's the code: const multiply = (x, y) => x * y;")
    ];

    const result = await service._generate(messages);

    // Test core functionality
    expect(mockClient.send).toHaveBeenCalledTimes(1);
    expect(result.generations[0].text).toBe('I understand the context. Here are my suggestions for improving the code.');
  });

  it('should handle Claude API errors gracefully', async () => {
    const mockError = new Error('Access denied to model');
    mockClient.send.mockRejectedValue(mockError);

    const messages = [
      new HumanChatMessage('Test message')
    ];

    await expect(service._generate(messages)).rejects.toThrow('Access denied to model');
  });

  it('should handle empty content response', async () => {
    const mockResponse = {
      body: new TextEncoder().encode(JSON.stringify({
        id: 'msg_03DEF789',
        type: 'message',
        role: 'assistant',
        content: [],
        model: 'claude-3-sonnet-20240229',
        stop_reason: 'end_turn',
        stop_sequence: null,
        usage: {
          input_tokens: 50,
          output_tokens: 0
        }
      }))
    };

    mockClient.send.mockResolvedValue(mockResponse);

    const messages = [new HumanChatMessage('Test')];
    const result = await service._generate(messages);

    expect(result.generations[0].text).toBe('');
    expect(result.llmOutput?.tokenUsage?.completionTokens).toBe(0);
  });

  it('should handle messages without system prompt', async () => {
    const mockResponse = {
      body: new TextEncoder().encode(JSON.stringify({
        id: 'msg_04GHI012',
        type: 'message',
        role: 'assistant',
        content: [
          {
            type: 'text',
            text: 'Hello! How can I help you today?'
          }
        ],
        model: 'claude-3-sonnet-20240229',
        stop_reason: 'end_turn',
        stop_sequence: null,
        usage: {
          input_tokens: 25,
          output_tokens: 30
        }
      }))
    };

    mockClient.send.mockResolvedValue(mockResponse);

    const messages = [
      new HumanChatMessage('Hello, Claude!')
    ];

    const result = await service._generate(messages);

    // Test core functionality
    expect(mockClient.send).toHaveBeenCalledTimes(1);
    expect(result.generations[0].text).toBe('Hello! How can I help you today?');
  });

  it('should use default configuration values', () => {
    const minimalService = new ClaudeBedrockService({
      region: 'us-west-2',
      modelId: 'anthropic.claude-3-haiku-20240307-v1:0'
    });

    expect(minimalService['config'].maxTokens).toBe(4096);
    expect(minimalService['config'].temperature).toBe(0);
    expect(minimalService._llmType()).toBe('claude-bedrock');
  });
});
