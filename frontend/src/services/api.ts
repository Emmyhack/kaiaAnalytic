import axios from 'axios';

const API_BASE_URL = process.env.REACT_APP_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    console.log(`Making ${config.method?.toUpperCase()} request to ${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    return Promise.reject(error);
  }
);

// Types
export interface HealthResponse {
  status: string;
  timestamp: string;
  version: string;
}

export interface BlockResponse {
  number: string;
  hash: string;
  parent_hash: string;
  timestamp: number;
  gas_used: number;
  gas_limit: number;
  transaction_count: number;
  size: string;
}

export interface TransactionResponse {
  hash: string;
  block_number?: string;
  block_hash?: string;
  transaction_index?: number;
  from: string;
  to?: string;
  value: string;
  gas: number;
  gas_price: string;
  gas_used?: number;
  status?: number;
}

export interface BalanceResponse {
  address: string;
  balance: string;
  balance_eth: string;
}

export interface NetworkStatsResponse {
  latest_block: number;
  network_id: string;
  chain_id: string;
  is_syncing: boolean;
  peer_count: number;
}

export interface ContractInfoResponse {
  address: string;
  code: string;
  code_size: number;
  is_contract: boolean;
}

export interface ErrorResponse {
  error: string;
  message: string;
}

// API functions
export const apiService = {
  // Health check
  async getHealth(): Promise<HealthResponse> {
    const response = await api.get<HealthResponse>('/health');
    return response.data;
  },

  // Block information
  async getBlock(blockNumber: string | number): Promise<BlockResponse> {
    const response = await api.get<BlockResponse>(`/api/v1/block/${blockNumber}`);
    return response.data;
  },

  // Transaction information
  async getTransaction(txHash: string): Promise<TransactionResponse> {
    const response = await api.get<TransactionResponse>(`/api/v1/transaction/${txHash}`);
    return response.data;
  },

  // Address balance
  async getBalance(address: string): Promise<BalanceResponse> {
    const response = await api.get<BalanceResponse>(`/api/v1/address/${address}/balance`);
    return response.data;
  },

  // Network statistics
  async getNetworkStats(): Promise<NetworkStatsResponse> {
    const response = await api.get<NetworkStatsResponse>('/api/v1/network/stats');
    return response.data;
  },

  // Contract information
  async getContractInfo(address: string): Promise<ContractInfoResponse> {
    const response = await api.get<ContractInfoResponse>(`/api/v1/contract/${address}/info`);
    return response.data;
  },
};

export default api;