import { useState, useEffect, createContext, useContext } from 'react';
import { subscriptionAPI } from '../services/api';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check for stored user address
    const storedAddress = localStorage.getItem('userAddress');
    if (storedAddress) {
      setUser({ address: storedAddress });
      checkSubscriptionStatus(storedAddress);
    }
    setLoading(false);
  }, []);

  const checkSubscriptionStatus = async (address) => {
    try {
      const response = await subscriptionAPI.getStatus(address);
      if (response.data.hasActiveSubscription) {
        setUser(prev => ({
          ...prev,
          subscription: response.data.subscription,
        }));
      }
    } catch (error) {
      console.error('Failed to check subscription status:', error);
    }
  };

  const connectWallet = async () => {
    try {
      // Mock wallet connection - in real implementation, connect to actual wallet
      const mockAddress = '0x' + Math.random().toString(16).substr(2, 40);
      
      setUser({ address: mockAddress });
      localStorage.setItem('userAddress', mockAddress);
      
      // Check subscription status
      await checkSubscriptionStatus(mockAddress);
      
      return mockAddress;
    } catch (error) {
      console.error('Failed to connect wallet:', error);
      throw error;
    }
  };

  const disconnectWallet = () => {
    setUser(null);
    localStorage.removeItem('userAddress');
  };

  const logout = () => {
    disconnectWallet();
  };

  const value = {
    user,
    loading,
    connectWallet,
    disconnectWallet,
    logout,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};