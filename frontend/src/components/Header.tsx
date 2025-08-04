import React, { useState, useEffect } from 'react';
import { Activity, Wifi, WifiOff } from 'lucide-react';
import { apiService, HealthResponse } from '../services/api';

const Header: React.FC = () => {
  const [health, setHealth] = useState<HealthResponse | null>(null);
  const [isOnline, setIsOnline] = useState(true);

  useEffect(() => {
    const checkHealth = async () => {
      try {
        const healthData = await apiService.getHealth();
        setHealth(healthData);
        setIsOnline(healthData.status === 'healthy');
      } catch (error) {
        setIsOnline(false);
        setHealth(null);
      }
    };

    checkHealth();
    const interval = setInterval(checkHealth, 30000); // Check every 30 seconds

    return () => clearInterval(interval);
  }, []);

  return (
    <header className="bg-white shadow-sm border-b border-secondary-200">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo and Title */}
          <div className="flex items-center space-x-3">
            <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-r from-primary-500 to-primary-600 rounded-lg">
              <Activity className="w-6 h-6 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-gradient">
                Kaia Analytics AI
              </h1>
              <p className="text-sm text-secondary-500">
                Blockchain Analytics Platform
              </p>
            </div>
          </div>

          {/* Status Indicator */}
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              {isOnline ? (
                <>
                  <Wifi className="w-5 h-5 text-green-500" />
                  <span className="text-sm text-green-600 font-medium">
                    Connected
                  </span>
                </>
              ) : (
                <>
                  <WifiOff className="w-5 h-5 text-red-500" />
                  <span className="text-sm text-red-600 font-medium">
                    Disconnected
                  </span>
                </>
              )}
            </div>

            {health && (
              <div className="hidden sm:flex items-center space-x-2 px-3 py-1 bg-secondary-50 rounded-full">
                <div className={`w-2 h-2 rounded-full ${
                  health.status === 'healthy' 
                    ? 'bg-green-500 animate-pulse' 
                    : 'bg-red-500'
                }`} />
                <span className="text-xs text-secondary-600">
                  API v{health.version}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  );
};

export default Header;