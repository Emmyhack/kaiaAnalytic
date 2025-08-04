import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Select, DatePicker, Button, Table, Tag, Tabs } from 'antd';
import { DownloadOutlined, FilterOutlined } from '@ant-design/icons';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ScatterChart,
  Scatter,
} from 'recharts';
import styled from 'styled-components';

const { RangePicker } = DatePicker;
const { Option } = Select;
const { TabPane } = Tabs;

const AnalyticsContainer = styled.div`
  .analytics-card {
    background: rgba(26, 26, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    margin-bottom: 24px;
  }
  
  .filters-section {
    background: rgba(26, 26, 46, 0.5);
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 24px;
  }
`;

// Mock data for different analytics
const yieldData = [
  { date: '2024-01-01', kaiaSwap: 12.5, kaiaLend: 8.2, kaiaStake: 6.8, tvl: 45200000 },
  { date: '2024-01-02', kaiaSwap: 12.8, kaiaLend: 8.1, kaiaStake: 6.9, tvl: 46100000 },
  { date: '2024-01-03', kaiaSwap: 12.3, kaiaLend: 8.3, kaiaStake: 6.7, tvl: 44800000 },
  { date: '2024-01-04', kaiaSwap: 13.1, kaiaLend: 8.0, kaiaStake: 7.0, tvl: 47500000 },
  { date: '2024-01-05', kaiaSwap: 12.9, kaiaLend: 8.4, kaiaStake: 6.8, tvl: 46800000 },
];

const tradingData = [
  { time: '00:00', volume: 1200000, price: 1.12, trades: 340 },
  { time: '04:00', volume: 1450000, price: 1.14, trades: 420 },
  { time: '08:00', volume: 1380000, price: 1.13, trades: 380 },
  { time: '12:00', volume: 1650000, price: 1.16, trades: 510 },
  { time: '16:00', volume: 1580000, price: 1.15, trades: 480 },
  { time: '20:00', volume: 1720000, price: 1.18, trades: 560 },
];

const governanceData = [
  {
    key: '1',
    proposal: 'KIP-001: Increase Block Gas Limit',
    category: 'Technical',
    sentiment: 0.75,
    participation: 0.68,
    status: 'Active',
    endDate: '2024-01-15',
  },
  {
    key: '2',
    proposal: 'KIP-002: Fee Structure Update',
    category: 'Economic',
    sentiment: 0.82,
    participation: 0.71,
    status: 'Passed',
    endDate: '2024-01-10',
  },
  {
    key: '3',
    proposal: 'KIP-003: Validator Rewards',
    category: 'Governance',
    sentiment: 0.65,
    participation: 0.59,
    status: 'Under Review',
    endDate: '2024-01-20',
  },
];

const Analytics: React.FC = () => {
  const [selectedTimeframe, setSelectedTimeframe] = useState('7d');
  const [selectedProtocol, setSelectedProtocol] = useState('all');
  const [loading, setLoading] = useState(false);

  const handleExportData = () => {
    // Implement data export functionality
    console.log('Exporting analytics data...');
  };

  const governanceColumns = [
    {
      title: 'Proposal',
      dataIndex: 'proposal',
      key: 'proposal',
      width: '30%',
    },
    {
      title: 'Category',
      dataIndex: 'category',
      key: 'category',
      render: (category: string) => (
        <Tag color={category === 'Technical' ? 'blue' : category === 'Economic' ? 'green' : 'purple'}>
          {category}
        </Tag>
      ),
    },
    {
      title: 'Sentiment',
      dataIndex: 'sentiment',
      key: 'sentiment',
      render: (sentiment: number) => (
        <span style={{ color: sentiment > 0.7 ? '#52c41a' : sentiment > 0.5 ? '#faad14' : '#ff4d4f' }}>
          {(sentiment * 100).toFixed(0)}%
        </span>
      ),
    },
    {
      title: 'Participation',
      dataIndex: 'participation',
      key: 'participation',
      render: (participation: number) => (
        <span>{(participation * 100).toFixed(1)}%</span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'Active' ? 'green' : status === 'Passed' ? 'blue' : 'orange'}>
          {status}
        </Tag>
      ),
    },
    {
      title: 'End Date',
      dataIndex: 'endDate',
      key: 'endDate',
    },
  ];

  return (
    <AnalyticsContainer>
      {/* Filters Section */}
      <div className="filters-section">
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={6}>
            <Select
              value={selectedTimeframe}
              onChange={setSelectedTimeframe}
              style={{ width: '100%' }}
              placeholder="Select timeframe"
            >
              <Option value="1d">Last 24 Hours</Option>
              <Option value="7d">Last 7 Days</Option>
              <Option value="30d">Last 30 Days</Option>
              <Option value="90d">Last 90 Days</Option>
            </Select>
          </Col>
          <Col xs={24} sm={6}>
            <Select
              value={selectedProtocol}
              onChange={setSelectedProtocol}
              style={{ width: '100%' }}
              placeholder="Select protocol"
            >
              <Option value="all">All Protocols</Option>
              <Option value="kaiaswap">KaiaSwap</Option>
              <Option value="kaialend">KaiaLend</Option>
              <Option value="kaiastake">KaiaStake</Option>
            </Select>
          </Col>
          <Col xs={24} sm={8}>
            <RangePicker style={{ width: '100%' }} />
          </Col>
          <Col xs={24} sm={4}>
            <Button
              type="primary"
              icon={<DownloadOutlined />}
              onClick={handleExportData}
              style={{ width: '100%' }}
            >
              Export
            </Button>
          </Col>
        </Row>
      </div>

      {/* Analytics Tabs */}
      <Tabs defaultActiveKey="yield" size="large">
        <TabPane tab="Yield Analytics" key="yield">
          <Row gutter={[24, 24]}>
            <Col xs={24} lg={16}>
              <Card title="APY Trends by Protocol" className="analytics-card">
                <ResponsiveContainer width="100%" height={400}>
                  <LineChart data={yieldData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="date" stroke="#a0a0a0" />
                    <YAxis stroke="#a0a0a0" />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: 'rgba(26, 26, 46, 0.9)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px',
                        color: '#ffffff'
                      }}
                    />
                    <Line 
                      type="monotone" 
                      dataKey="kaiaSwap" 
                      stroke="#667eea" 
                      strokeWidth={3}
                      name="KaiaSwap APY"
                    />
                    <Line 
                      type="monotone" 
                      dataKey="kaiaLend" 
                      stroke="#52c41a" 
                      strokeWidth={3}
                      name="KaiaLend APY"
                    />
                    <Line 
                      type="monotone" 
                      dataKey="kaiaStake" 
                      stroke="#faad14" 
                      strokeWidth={3}
                      name="KaiaStake APY"
                    />
                  </LineChart>
                </ResponsiveContainer>
              </Card>
            </Col>
            <Col xs={24} lg={8}>
              <Card title="Total Value Locked (TVL)" className="analytics-card">
                <ResponsiveContainer width="100%" height={400}>
                  <AreaChart data={yieldData}>
                    <defs>
                      <linearGradient id="colorTVL" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#764ba2" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#764ba2" stopOpacity={0.1}/>
                      </linearGradient>
                    </defs>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="date" stroke="#a0a0a0" />
                    <YAxis stroke="#a0a0a0" />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: 'rgba(26, 26, 46, 0.9)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px',
                        color: '#ffffff'
                      }}
                      formatter={(value: number) => [`$${(value / 1000000).toFixed(1)}M`, 'TVL']}
                    />
                    <Area 
                      type="monotone" 
                      dataKey="tvl" 
                      stroke="#764ba2" 
                      fillOpacity={1} 
                      fill="url(#colorTVL)" 
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="Trading Analytics" key="trading">
          <Row gutter={[24, 24]}>
            <Col xs={24} lg={12}>
              <Card title="Trading Volume" className="analytics-card">
                <ResponsiveContainer width="100%" height={350}>
                  <BarChart data={tradingData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis dataKey="time" stroke="#a0a0a0" />
                    <YAxis stroke="#a0a0a0" />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: 'rgba(26, 26, 46, 0.9)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px',
                        color: '#ffffff'
                      }}
                      formatter={(value: number) => [`$${(value / 1000000).toFixed(1)}M`, 'Volume']}
                    />
                    <Bar 
                      dataKey="volume" 
                      fill="url(#colorVolume)" 
                      radius={[4, 4, 0, 0]}
                    />
                    <defs>
                      <linearGradient id="colorVolume" x1="0" y1="0" x2="0" y2="1">
                        <stop offset="5%" stopColor="#667eea" stopOpacity={0.8}/>
                        <stop offset="95%" stopColor="#667eea" stopOpacity={0.3}/>
                      </linearGradient>
                    </defs>
                  </BarChart>
                </ResponsiveContainer>
              </Card>
            </Col>
            <Col xs={24} lg={12}>
              <Card title="Price vs Volume Correlation" className="analytics-card">
                <ResponsiveContainer width="100%" height={350}>
                  <ScatterChart data={tradingData}>
                    <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.1)" />
                    <XAxis 
                      type="number" 
                      dataKey="volume" 
                      name="Volume" 
                      stroke="#a0a0a0"
                      tickFormatter={(value) => `$${(value / 1000000).toFixed(1)}M`}
                    />
                    <YAxis 
                      type="number" 
                      dataKey="price" 
                      name="Price" 
                      stroke="#a0a0a0"
                      tickFormatter={(value) => `$${value.toFixed(2)}`}
                    />
                    <Tooltip 
                      contentStyle={{
                        backgroundColor: 'rgba(26, 26, 46, 0.9)',
                        border: '1px solid rgba(255, 255, 255, 0.1)',
                        borderRadius: '8px',
                        color: '#ffffff'
                      }}
                      formatter={(value, name) => [
                        name === 'Volume' ? `$${(Number(value) / 1000000).toFixed(1)}M` : `$${Number(value).toFixed(3)}`,
                        name
                      ]}
                    />
                    <Scatter dataKey="trades" fill="#52c41a" />
                  </ScatterChart>
                </ResponsiveContainer>
              </Card>
            </Col>
          </Row>
        </TabPane>

        <TabPane tab="Governance Analytics" key="governance">
          <Row gutter={[24, 24]}>
            <Col span={24}>
              <Card title="Governance Proposals Analysis" className="analytics-card">
                <Table
                  columns={governanceColumns}
                  dataSource={governanceData}
                  pagination={false}
                  size="middle"
                />
              </Card>
            </Col>
          </Row>
        </TabPane>
      </Tabs>
    </AnalyticsContainer>
  );
};

export default Analytics;