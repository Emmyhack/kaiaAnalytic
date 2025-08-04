import React, { useState, useEffect } from 'react';
import { useAccount, useBalance } from 'wagmi';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, LineChart, Line } from 'recharts';
import { TrendingUp, TrendingDown, DollarSign, Wallet, RefreshCw } from 'lucide-react';
import { formatEther, formatUSD } from '../utils/format';

interface AssetData {
  name: string;
  symbol: string;
  balance: string;
  value: number;
  change24h: number;
  color: string;
}

interface PortfolioMetrics {
  totalValue: number;
  change24h: number;
  change7d: number;
  change30d: number;
}

const COLORS = ['#3B82F6', '#10B981', '#F59E0B', '#EF4444', '#8B5CF6', '#06B6D4'];

const PortfolioTracker: React.FC = () => {
  const { address, isConnected } = useAccount();
  const { data: ethBalance, refetch: refetchBalance } = useBalance({ address });
  
  const [assets, setAssets] = useState<AssetData[]>([]);
  const [metrics, setMetrics] = useState<PortfolioMetrics>({
    totalValue: 0,
    change24h: 0,
    change7d: 0,
    change30d: 0
  });
  const [loading, setLoading] = useState(false);
  const [timeRange, setTimeRange] = useState<'24h' | '7d' | '30d' | '1y'>('7d');

  // Mock historical data for demonstration
  const mockHistoricalData = [
    { date: '2025-07-01', value: 12500 },
    { date: '2025-07-08', value: 13200 },
    { date: '2025-07-15', value: 12800 },
    { date: '2025-07-22', value: 14100 },
    { date: '2025-07-29', value: 15300 },
    { date: '2025-08-04', value: 16750 },
  ];

  useEffect(() => {
    if (isConnected && address && ethBalance) {
      loadPortfolioData();
    }
  }, [isConnected, address, ethBalance]);

  const loadPortfolioData = async () => {
    setLoading(true);
    try {
      // Mock portfolio data - in a real app, this would come from your API
      const mockAssets: AssetData[] = [
        {
          name: 'Ethereum',
          symbol: 'ETH',
          balance: ethBalance ? formatEther(ethBalance.value.toString()) : '0',
          value: parseFloat(ethBalance ? formatEther(ethBalance.value.toString()) : '0') * 3200, // Mock ETH price
          change24h: 2.5,
          color: COLORS[0]
        },
        {
          name: 'USD Coin',
          symbol: 'USDC',
          balance: '1250.00',
          value: 1250,
          change24h: 0.1,
          color: COLORS[1]
        },
        {
          name: 'Wrapped Bitcoin',
          symbol: 'WBTC',
          balance: '0.125',
          value: 0.125 * 65000, // Mock BTC price
          change24h: -1.2,
          color: COLORS[2]
        }
      ];

      setAssets(mockAssets);
      
      const totalValue = mockAssets.reduce((sum, asset) => sum + asset.value, 0);
      setMetrics({
        totalValue,
        change24h: 3.2,
        change7d: 12.8,
        change30d: 28.5
      });
    } catch (error) {
      console.error('Failed to load portfolio data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    refetchBalance();
    loadPortfolioData();
  };

  if (!isConnected) {
    return (
      <div className="card">
        <div className="text-center py-12">
          <Wallet className="w-16 h-16 text-secondary-300 mx-auto mb-4" />
          <h3 className="text-lg font-semibold text-secondary-900 mb-2">
            Connect Your Wallet
          </h3>
          <p className="text-secondary-600 mb-6">
            Connect your wallet to view your portfolio and track your assets.
          </p>
          <button 
            onClick={() => (window as any).w3m?.open?.()}
            className="btn-primary"
          >
            <Wallet className="w-4 h-4 mr-2" />
            Connect Wallet
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Portfolio Header */}
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold text-secondary-900">Portfolio</h2>
        <button
          onClick={handleRefresh}
          disabled={loading}
          className="btn-secondary"
        >
          <RefreshCw className={`w-4 h-4 mr-2 ${loading ? 'animate-spin' : ''}`} />
          Refresh
        </button>
      </div>

      {/* Portfolio Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        <div className="card">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-secondary-600">Total Value</span>
            <DollarSign className="w-4 h-4 text-secondary-400" />
          </div>
          <div className="text-2xl font-bold text-secondary-900">
            {formatUSD(metrics.totalValue)}
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-secondary-600">24h Change</span>
            {metrics.change24h >= 0 ? (
              <TrendingUp className="w-4 h-4 text-green-500" />
            ) : (
              <TrendingDown className="w-4 h-4 text-red-500" />
            )}
          </div>
          <div className={`text-2xl font-bold ${
            metrics.change24h >= 0 ? 'text-green-600' : 'text-red-600'
          }`}>
            {metrics.change24h >= 0 ? '+' : ''}{metrics.change24h.toFixed(2)}%
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-secondary-600">7d Change</span>
            {metrics.change7d >= 0 ? (
              <TrendingUp className="w-4 h-4 text-green-500" />
            ) : (
              <TrendingDown className="w-4 h-4 text-red-500" />
            )}
          </div>
          <div className={`text-2xl font-bold ${
            metrics.change7d >= 0 ? 'text-green-600' : 'text-red-600'
          }`}>
            {metrics.change7d >= 0 ? '+' : ''}{metrics.change7d.toFixed(2)}%
          </div>
        </div>

        <div className="card">
          <div className="flex items-center justify-between mb-2">
            <span className="text-sm font-medium text-secondary-600">30d Change</span>
            {metrics.change30d >= 0 ? (
              <TrendingUp className="w-4 h-4 text-green-500" />
            ) : (
              <TrendingDown className="w-4 h-4 text-red-500" />
            )}
          </div>
          <div className={`text-2xl font-bold ${
            metrics.change30d >= 0 ? 'text-green-600' : 'text-red-600'
          }`}>
            {metrics.change30d >= 0 ? '+' : ''}{metrics.change30d.toFixed(2)}%
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Asset Allocation Chart */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-semibold text-secondary-900">Asset Allocation</h3>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={assets}
                  cx="50%"
                  cy="50%"
                  outerRadius={80}
                  dataKey="value"
                  label={({ name, percent }) => `${name} ${((percent || 0) * 100).toFixed(1)}%`}
                >
                  {assets.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip formatter={(value: number) => formatUSD(value)} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Portfolio Performance Chart */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-semibold text-secondary-900">Portfolio Performance</h3>
            <div className="flex space-x-2">
              {(['24h', '7d', '30d', '1y'] as const).map((range) => (
                <button
                  key={range}
                  onClick={() => setTimeRange(range)}
                  className={`px-3 py-1 text-sm rounded-md transition-colors ${
                    timeRange === range
                      ? 'bg-primary-100 text-primary-700'
                      : 'text-secondary-600 hover:text-secondary-900'
                  }`}
                >
                  {range}
                </button>
              ))}
            </div>
          </div>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <LineChart data={mockHistoricalData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip formatter={(value: number) => formatUSD(value)} />
                <Line 
                  type="monotone" 
                  dataKey="value" 
                  stroke="#3B82F6" 
                  strokeWidth={2}
                  dot={{ fill: '#3B82F6' }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Asset List */}
      <div className="card">
        <div className="card-header">
          <h3 className="text-lg font-semibold text-secondary-900">Assets</h3>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full">
            <thead>
              <tr className="border-b border-secondary-200">
                <th className="text-left py-3 px-4 font-medium text-secondary-600">Asset</th>
                <th className="text-right py-3 px-4 font-medium text-secondary-600">Balance</th>
                <th className="text-right py-3 px-4 font-medium text-secondary-600">Value</th>
                <th className="text-right py-3 px-4 font-medium text-secondary-600">24h Change</th>
              </tr>
            </thead>
            <tbody>
              {assets.map((asset, index) => (
                <tr key={index} className="border-b border-secondary-100 hover:bg-secondary-50">
                  <td className="py-3 px-4">
                    <div className="flex items-center space-x-3">
                      <div 
                        className="w-8 h-8 rounded-full flex items-center justify-center text-white font-semibold text-sm"
                        style={{ backgroundColor: asset.color }}
                      >
                        {asset.symbol.charAt(0)}
                      </div>
                      <div>
                        <div className="font-medium text-secondary-900">{asset.name}</div>
                        <div className="text-sm text-secondary-500">{asset.symbol}</div>
                      </div>
                    </div>
                  </td>
                  <td className="text-right py-3 px-4 font-mono text-secondary-900">
                    {asset.balance}
                  </td>
                  <td className="text-right py-3 px-4 font-semibold text-secondary-900">
                    {formatUSD(asset.value)}
                  </td>
                  <td className={`text-right py-3 px-4 font-semibold ${
                    asset.change24h >= 0 ? 'text-green-600' : 'text-red-600'
                  }`}>
                    {asset.change24h >= 0 ? '+' : ''}{asset.change24h.toFixed(2)}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
};

export default PortfolioTracker;