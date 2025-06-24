import request from './request';

// Claude API 接口的类型定义
export interface ClaudeMessage {
  id: string;
  type: 'human' | 'assistant';
  role?: string;
  content: Array<{
    type: 'text' | 'image' | 'tool_use' | 'tool_result';
    text?: string;
    source?: {
      type: string;
      media_type: string;
      data: string;
    };
    tool_use?: {
      id: string;
      name: string;
      input: any;
    };
    tool_result?: {
      tool_use_id: string;
      content: string;
    };
  }>;
}

export interface ClaudeRequestBody {
  model: string;
  messages: ClaudeMessage[];
  system?: string;
  max_tokens?: number;
  temperature?: number;
  top_p?: number;
  top_k?: number;
  tools?: Array<{
    name: string;
    description: string;
    input_schema: any;
  }>;
  tool_choice?: string | {
    type: string;
    name?: string;
  };
  stop_sequences?: string[];
  stream?: boolean;
}

export interface ClaudeResponse {
  id: string;
  type: 'message';
  role: 'assistant';
  model: string;
  content: Array<{
    type: string;
    text?: string;
    tool_use?: any;
  }>;
  usage: {
    input_tokens: number;
    output_tokens: number;
  };
  stop_reason: string | null;
  stop_sequence: string | null;
  tool_use?: any;
}

/**
 * 发送请求到Claude API
 */
export function sendClaudeRequest(data: ClaudeRequestBody): Promise<ClaudeResponse> {
  return request({
    url: '/claude',
    method: 'post',
    data,
  });
}

/**
 * 发送请求到Claude API的任意端点
 * @param path 子路径
 * @param method HTTP方法
 * @param data 请求数据
 * @param params 查询参数
 */
export function sendClaudePathRequest(
  path: string,
  method: 'get' | 'post' | 'put' | 'delete',
  data?: any,
  params?: any
): Promise<any> {
  return request({
    url: `/claude/${path}`,
    method,
    data,
    params,
  });
}