export const dynamic = 'force-dynamic'

import { NextRequest, NextResponse } from 'next/server';

// 动态获取API_URL，支持环境变量配置
const getApiBaseUrl = () => {
  const apiUrl = process.env.API_URL;
  if (!apiUrl) {
    console.warn('API_URL environment variable not set, using default');
    return 'http://localhost:9998';
  }
  return apiUrl;
};

export async function GET(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams.path, 'GET');
}

export async function POST(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams.path, 'POST');
}

export async function PUT(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams.path, 'PUT');
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams.path, 'DELETE');
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const resolvedParams = await params;
  return proxyRequest(request, resolvedParams.path, 'PATCH');
}

async function proxyRequest(
  request: NextRequest,
  pathSegments: string[],
  method: string
) {
  try {
    const path = pathSegments.join('/');
    const url = new URL(request.url);
    const queryString = url.search;
    
    const API_BASE_URL = getApiBaseUrl();
    const targetUrl = `${API_BASE_URL}/${path}${queryString}`;
    
    console.log(`[API Proxy] ${method} ${targetUrl}`);
    
    // 准备请求头，排除一些不需要的头
    const headers = new Headers();
    request.headers.forEach((value, key) => {
      // 排除一些不应该转发的头
      if (!['host', 'connection', 'content-length'].includes(key.toLowerCase())) {
        headers.set(key, value);
      }
    });
    
    // 准备请求体
    let body: BodyInit | null = null;
    if (['POST', 'PUT', 'PATCH'].includes(method)) {
      body = await request.text();
    }
    
    // 发送代理请求
    const response = await fetch(targetUrl, {
      method,
      headers,
      body,
    });
    
    // 准备响应头
    const responseHeaders = new Headers();
    response.headers.forEach((value, key) => {
      // 排除一些不应该转发的响应头
      if (!['content-encoding', 'transfer-encoding'].includes(key.toLowerCase())) {
        responseHeaders.set(key, value);
      }
    });
    
    // 获取响应内容
    const responseBody = await response.arrayBuffer();
    
    return new NextResponse(responseBody, {
      status: response.status,
      statusText: response.statusText,
      headers: responseHeaders,
    });
    
  } catch (error) {
    console.error('[API Proxy] Error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error', message: 'Failed to proxy request' },
      { status: 500 }
    );
  }
}