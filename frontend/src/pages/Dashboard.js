import React from 'react';
import { useQuery } from 'react-query';
import { format } from 'date-fns';
import {
  TrendingUpIcon,
  TrendingDownIcon,
  CurrencyDollarIcon,
  ChartBarIcon,
  ChatBubbleLeftRightIcon,
  ClockIcon,
} from '@heroicons/react/24/outline';
import ReactECharts from 'echarts-for-react';
import { api } from '../services/api';

function Dashboard() {
  // Fetch dashboard data
  const { data: analyticsData, isLoading: analyticsLoading } = useQuery(
    'analytics',
    () => api.get('/api/v1/analytics/yield'),
    { refetchInterval: 30000 }
  );

  const { data: volumeData, isLoading: volumeLoading } = useQuery(
    'volume',
    () => api.get('/api/v1/analytics/volume'),
    { refetchInterval: 30000 }
  );

  const { data: gasData, isLoading: gasLoading } = useQuery(
    'gas',
    () => api.get('/api/v1/analytics/gas'),
    { refetchInterval: 30000 }
  );

  // Mock data for demonstration
  const mockData = {
    totalVolume: 1800000,
    volumeChange: 12.5,
    activeUsers: 15000,
    usersChange: 8.2,
    gasPrice: 25,
    gasChange: -2.1,
    yieldOpportunities: 5,
    yieldChange: 15.3,
  };

  // Chart options
  const volumeChartOption = {
    title: {
      text: 'Transaction Volume (24h)',
      left: 'center',
      textStyle: {
        fontSize: 14,
        fontWeight: 'normal',
      },
    },
    tooltip: {
      trigger: 'axis',
    },
    xAxis: {
      type: 'category',
      data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'],
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: '{value} K',
      },
    },
    series: [
      {
        name: 'Volume',
        type: 'line',
        smooth: true,
        data: [1200, 1350, 1100, 1600, 1800, 1700, 1500],
        itemStyle: {
          color: '#3B82F6',
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(59, 130, 246, 0.3)' },
              { offset: 1, color: 'rgba(59, 130, 246, 0.1)' },
            ],
          },
        },
      },
    ],
  };

  const gasChartOption = {
    title: {
      text: 'Gas Price Trend',
      left: 'center',
      textStyle: {
        fontSize: 14,
        fontWeight: 'normal',
      },
    },
    tooltip: {
      trigger: 'axis',
    },
    xAxis: {
      type: 'category',
      data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00', '24:00'],
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        formatter: '{value} Gwei',
      },
    },
    series: [
      {
        name: 'Gas Price',
        type: 'line',
        smooth: true,
        data: [22, 25, 28, 30, 27, 25, 23],
        itemStyle: {
          color: '#10B981',
        },
      },
    ],
  };

  const yieldChartOption = {
    title: {
      text: 'Top Yield Opportunities',
      left: 'center',
      textStyle: {
        fontSize: 14,
        fontWeight: 'normal',
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}% APY',
    },
    series: [
      {
        name: 'APY',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2,
        },
        label: {
          show: false,
          position: 'center',
        },
        emphasis: {
          label: {
            show: true,
            fontSize: '18',
            fontWeight: 'bold',
          },
        },
        labelLine: {
          show: false,
        },
        data: [
          { value: 12.5, name: 'KAIA-USDC LP' },
          { value: 8.2, name: 'KAIA-ETH LP' },
          { value: 6.8, name: 'KAIA-BTC LP' },
          { value: 5.5, name: 'Staking Pool' },
        ],
      },
    ],
  };

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div>
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-1 text-sm text-gray-500">
          Welcome to KaiaAnalyticsAI. Here's your overview of the Kaia ecosystem.
        </p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <CurrencyDollarIcon className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Total Volume (24h)
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      ${(mockData.totalVolume / 1000000).toFixed(1)}M
                    </div>
                    <div className="ml-2 flex items-baseline text-sm font-semibold text-green-600">
                      <TrendingUpIcon className="self-center flex-shrink-0 h-4 w-4 text-green-500" />
                      <span className="sr-only">Increased</span>
                      {mockData.volumeChange}%
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <ChartBarIcon className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Active Users
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      {mockData.activeUsers.toLocaleString()}
                    </div>
                    <div className="ml-2 flex items-baseline text-sm font-semibold text-green-600">
                      <TrendingUpIcon className="self-center flex-shrink-0 h-4 w-4 text-green-500" />
                      <span className="sr-only">Increased</span>
                      {mockData.usersChange}%
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <ClockIcon className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Gas Price
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      {mockData.gasPrice} Gwei
                    </div>
                    <div className="ml-2 flex items-baseline text-sm font-semibold text-red-600">
                      <TrendingDownIcon className="self-center flex-shrink-0 h-4 w-4 text-red-500" />
                      <span className="sr-only">Decreased</span>
                      {mockData.gasChange}%
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>

        <div className="bg-white overflow-hidden shadow rounded-lg">
          <div className="p-5">
            <div className="flex items-center">
              <div className="flex-shrink-0">
                <ChatBubbleLeftRightIcon className="h-6 w-6 text-gray-400" />
              </div>
              <div className="ml-5 w-0 flex-1">
                <dl>
                  <dt className="text-sm font-medium text-gray-500 truncate">
                    Yield Opportunities
                  </dt>
                  <dd className="flex items-baseline">
                    <div className="text-2xl font-semibold text-gray-900">
                      {mockData.yieldOpportunities}
                    </div>
                    <div className="ml-2 flex items-baseline text-sm font-semibold text-green-600">
                      <TrendingUpIcon className="self-center flex-shrink-0 h-4 w-4 text-green-500" />
                      <span className="sr-only">Increased</span>
                      {mockData.yieldChange}%
                    </div>
                  </dd>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="bg-white shadow rounded-lg p-6">
          <ReactECharts option={volumeChartOption} style={{ height: '300px' }} />
        </div>

        <div className="bg-white shadow rounded-lg p-6">
          <ReactECharts option={gasChartOption} style={{ height: '300px' }} />
        </div>
      </div>

      {/* Yield opportunities */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Yield Opportunities</h3>
        </div>
        <div className="p-6">
          <ReactECharts option={yieldChartOption} style={{ height: '300px' }} />
        </div>
      </div>

      {/* Recent activity */}
      <div className="bg-white shadow rounded-lg">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Recent Activity</h3>
        </div>
        <div className="divide-y divide-gray-200">
          {[
            {
              id: 1,
              type: 'yield',
              message: 'New yield opportunity detected: KAIA-USDC LP at 12.5% APY',
              time: new Date(),
            },
            {
              id: 2,
              type: 'governance',
              message: 'New governance proposal: Increase protocol fee to 0.15%',
              time: new Date(Date.now() - 3600000),
            },
            {
              id: 3,
              type: 'trading',
              message: 'Trading signal: Buy KAIA/USDC at $1.25',
              time: new Date(Date.now() - 7200000),
            },
          ].map((activity) => (
            <div key={activity.id} className="px-6 py-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="h-8 w-8 rounded-full bg-blue-100 flex items-center justify-center">
                      <span className="text-sm font-medium text-blue-600">
                        {activity.type.charAt(0).toUpperCase()}
                      </span>
                    </div>
                  </div>
                  <div className="ml-4">
                    <p className="text-sm font-medium text-gray-900">{activity.message}</p>
                    <p className="text-sm text-gray-500">
                      {format(activity.time, 'MMM d, yyyy HH:mm')}
                    </p>
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

export default Dashboard;