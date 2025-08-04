import React, { useState, useEffect } from 'react';
import { 
  Activity, 
  Blocks, 
  Network, 
  Users, 
  RefreshCw,
  TrendingUp,
  Clock
} from 'lucide-react';
import { apiService, NetworkStatsResponse } from '../services/api';
import { formatNumber, formatBlockNumber } from '../utils/format';

const NetworkStats: React.FC = () => {
  const [stats, setStats] = useState<NetworkStatsResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

  const fetchNetworkStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await apiService.getNetworkStats();
      setStats(data);
      setLastUpdated(new Date());
    } catch (err) {
      setError('Failed to fetch network statistics');
      console.error('Network stats error:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchNetworkStats();
    
    // Auto-refresh every 30 seconds
    const interval = setInterval(fetchNetworkStats, 30000);
    
    return () => clearInterval(interval);
  }, []);

  const getNetworkName = (chainId: string): string => {
    switch (chainId) {
      case '1':
        return 'Ethereum Mainnet';
      case '11155111':
        return 'Sepolia Testnet';
      case '137':
        return 'Polygon Mainnet';
      case '1337':
        return 'Local Network';
      default:
        return `Chain ID ${chainId}`;
    }
  };

  const StatCard: React.FC<{
    title: string;
    value: string | number;
    icon: React.ReactNode;
    subtitle?: string;
    trend?: 'up' | 'down' | 'neutral';
    color?: 'blue' | 'green' | 'purple' | 'orange';
  }> = ({ title, value, icon, subtitle, trend, color = 'blue' }) => {
    const colorClasses = {
      blue: 'from-blue-500 to-blue-600',
      green: 'from-green-500 to-green-600',
      purple: 'from-purple-500 to-purple-600',
      orange: 'from-orange-500 to-orange-600',
    };

    return (
      <div className="card">
        <div className="flex items-center justify-between">
          <div className="flex-1">
            <div className="flex items-center space-x-2 mb-2">
              <div className={`p-2 rounded-lg bg-gradient-to-r ${colorClasses[color]} text-white`}>
                {icon}
              </div>
              <h3 className="text-sm font-medium text-secondary-600">{title}</h3>
            </div>
            <div className="flex items-baseline space-x-2">
              <p className="text-2xl font-bold text-secondary-900">{value}</p>
              {trend && (
                <div className={`flex items-center text-sm ${
                  trend === 'up' ? 'text-green-600' : 
                  trend === 'down' ? 'text-red-600' : 'text-secondary-500'
                }`}>
                  <TrendingUp className={`w-4 h-4 ${trend === 'down' ? 'rotate-180' : ''}`} />
                </div>
              )}
            </div>
            {subtitle && (
              <p className="text-sm text-secondary-500 mt-1">{subtitle}</p>
            )}
          </div>
        </div>
      </div>
    );
  };

  if (loading && !stats) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="card animate-pulse">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-secondary-200 rounded-lg"></div>
              <div className="flex-1">
                <div className="h-4 bg-secondary-200 rounded w-24 mb-2"></div>
                <div className="h-6 bg-secondary-200 rounded w-16"></div>
              </div>
            </div>
          </div>
        ))}
      </div>
    );
  }

  if (error) {
    return (
      <div className="card">
        <div className="text-center py-8">
          <Activity className="w-12 h-12 text-red-400 mx-auto mb-4" />
          <h3 className="text-lg font-medium text-secondary-900 mb-2">
            Unable to load network statistics
          </h3>
          <p className="text-secondary-500 mb-4">{error}</p>
          <button 
            onClick={fetchNetworkStats}
            className="btn-primary"
            disabled={loading}
          >
            {loading ? (
              <div className="loading-spinner w-4 h-4 mr-2" />
            ) : (
              <RefreshCw className="w-4 h-4 mr-2" />
            )}
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!stats) return null;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-secondary-900">Network Statistics</h2>
        <div className="flex items-center space-x-4">
          {lastUpdated && (
            <div className="flex items-center text-sm text-secondary-500">
              <Clock className="w-4 h-4 mr-1" />
              Updated {lastUpdated.toLocaleTimeString()}
            </div>
          )}
          <button
            onClick={fetchNetworkStats}
            disabled={loading}
            className="btn-secondary"
          >
            <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <StatCard
          title="Latest Block"
          value={formatBlockNumber(stats.latest_block)}
          icon={<Blocks className="w-5 h-5" />}
          subtitle="Current block height"
          trend="up"
          color="blue"
        />

        <StatCard
          title="Network"
          value={getNetworkName(stats.chain_id)}
          icon={<Network className="w-5 h-5" />}
          subtitle={`Chain ID: ${stats.chain_id}`}
          color="green"
        />

        <StatCard
          title="Network ID"
          value={stats.network_id}
          icon={<Activity className="w-5 h-5" />}
          subtitle="Network identifier"
          color="purple"
        />

        <StatCard
          title="Sync Status"
          value={stats.is_syncing ? "Syncing" : "Synced"}
          icon={<Users className="w-5 h-5" />}
          subtitle={stats.is_syncing ? "Node synchronizing" : "Fully synchronized"}
          color={stats.is_syncing ? "orange" : "green"}
        />
      </div>

      {/* Additional Info */}
      <div className="card">
        <div className="card-header">
          <h3 className="text-lg font-semibold text-secondary-900">Network Information</h3>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Chain ID
            </label>
            <p className="text-lg font-mono text-secondary-900">{stats.chain_id}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Network ID
            </label>
            <p className="text-lg font-mono text-secondary-900">{stats.network_id}</p>
          </div>
          <div>
            <label className="block text-sm font-medium text-secondary-700 mb-1">
              Peer Count
            </label>
            <p className="text-lg font-mono text-secondary-900">
              {stats.peer_count > 0 ? formatNumber(stats.peer_count) : 'N/A'}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default NetworkStats;