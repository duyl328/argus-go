/**
 * Time:2025/6/15 20:12 37
 * Name:http.ts
 * Path:src/types
 * ProjectName:argus-front
 * Author:charlatans
 *
 *  Il n'ya qu'un héroïsme au monde :
 *     c'est de voir le monde tel qu'il est et de l'aimer.
 */

// 1. 配置接口和类型定义
export interface HttpConfig {
  baseURL: string;
  timeout?: number;
  headers?: Record<string, string>;
}

export interface ApiResponse<T = unknown> {
  code: number;
  data: T;
  message: string;
  success: boolean;
}

export interface RequestConfig {
  url: string;
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  params?: unknown;
  data?: unknown;
  headers?: Record<string, string>;
  timeout?: number;
}
