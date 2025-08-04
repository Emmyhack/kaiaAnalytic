import axios from 'axios';

// Create axios instance
export const api = axios.create({
  baseURL: process.env.REACT_APP_API_URL || 'http://localhost:8080',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add user address to headers if available
    const userAddress = localStorage.getItem('userAddress');
    if (userAddress) {
      config.headers['X-User-Address'] = userAddress;
    }
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
    // Handle common errors
    if (error.response?.status === 401) {
      // Handle unauthorized
      localStorage.removeItem('userAddress');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// API functions
export const analyticsAPI = {
  // Yield opportunities
  getYieldOpportunities: () => api.get('/api/v1/analytics/yield'),
  
  // Governance sentiment
  getGovernanceSentiment: () => api.get('/api/v1/analytics/governance'),
  
  // Trading suggestions
  getTradingSuggestions: () => api.get('/api/v1/analytics/trading'),
  
  // Transaction volume
  getTransactionVolume: () => api.get('/api/v1/analytics/volume'),
  
  // Gas trends
  getGasTrends: () => api.get('/api/v1/analytics/gas'),
};

export const collectorAPI = {
  // Blockchain data
  getBlockchainData: () => api.get('/api/v1/collector/blockchain'),
  
  // Market data
  getMarketData: () => api.get('/api/v1/collector/market'),
  
  // Historical data
  getHistoricalData: () => api.get('/api/v1/collector/historical'),
};

export const chatAPI = {
  // Send query
  sendQuery: (query) => api.post('/api/v1/chat/query', { query }),
};

export const subscriptionAPI = {
  // Get subscription plans
  getPlans: () => api.get('/api/v1/subscription/plans'),
  
  // Purchase subscription
  purchaseSubscription: (planId, userAddress) => 
    api.post('/api/v1/subscription/purchase', { planId, userAddress }),
  
  // Get subscription status
  getStatus: (address) => api.get(`/api/v1/subscription/status/${address}`),
};

export default api;