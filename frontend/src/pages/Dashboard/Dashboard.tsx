import React, { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, Progress, Table, Tag, Button } from 'antd';
import {
  ArrowUpOutlined,
  ArrowDownOutlined,
  DollarOutlined,
  TrendingUpOutlined,
  UserOutlined,
  FireOutlined,
} from '@ant-design/icons';
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
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import styled from 'styled-components';

const DashboardContainer = styled.div`
  .dashboard-card {
    background: rgba(26, 26, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
  }
  
  .metric-card {
    text-align: center;
    padding: 20px;
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    border: 1px solid rgba(102, 126, 234, 0.3);
    border-radius: 12px;
  }
  
  .trending-up {
    color: #52c41a;
  }
  
  .trending-down {
    color: #ff4d4f;
  }
`;

const priceData = [
  { time: '00:00', price: 1.12, volume: 1200000 },
  { time: '04:00', price: 1.14, volume: 1450000 },
  { time: '08:00', price: 1.13, volume: 1380000 },
  { time: '12:00', price: 1.16, volume: 1650000 },
  { time: '16:00', price: 1.15, volume: 1580000 },
  { time: '20:00', price: 1.18, volume: 1720000 },
];

const yieldData = [
  { protocol: 'KaiaSwap', apy: 12.5, tvl: 15600000, risk: 'Low' },
  { protocol: 'KaiaLend', apy: 8.2, tvl: 25400000, risk: 'Very Low' },
  { protocol: 'KaiaStake', apy: 6.8, tvl: 45200000, risk: 'Minimal' },
  { protocol: 'KaiaFarm', apy: 15.7, tvl: 8900000, risk: 'Medium' },
];

const portfolioData = [
  { name: 'KAIA', value: 45, color: '#667eea' },
  { name: 'USDC', value: 25, color: '#764ba2' },
  { name: 'ETH', value: 20, color: '#f093fb' },
  { name: 'BTC', value: 10, color: '#f5576c' },
];

const recentTrades = [
  {
    key: '1',
    pair: 'KAIA/USDC',
    type: 'Buy',
    amount: '1,000 KAIA',
    price: '$1.15',
    time: '2 min ago',
    status: 'Completed',
  },
  {
    key: '2',
    pair: 'ETH/KAIA',
    type: 'Sell',
    amount: '0.5 ETH',
    price: '1,750 KAIA',
    time: '15 min ago',
    status: 'Completed',
  },
  {
    key: '3',
    pair: 'KAIA/USDC',
    type: 'Buy',
    amount: '500 KAIA',
    price: '$1.14',
    time: '1 hour ago',
    status: 'Pending',
  },
];

const Dashboard: React.FC = () => {
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Simulate loading
    const timer = setTimeout(() => setLoading(false), 1000);
    return () => clearTimeout(timer);
  }, []);

  const tradeColumns = [
    {
      title: 'Pair',
      dataIndex: 'pair',
      key: 'pair',
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === 'Buy' ? 'green' : 'red'}>
          {type}
        </Tag>
      ),
    },
    {
      title: 'Amount',
      dataIndex: 'amount',
      key: 'amount',
    },
    {
      title: 'Price',
      dataIndex: 'price',
      key: 'price',
    },
    {
      title: 'Time',
      dataIndex: 'time',
      key: 'time',
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color={status === 'Completed' ? 'green' : 'orange'}>
          {status}
        </Tag>
      ),
    },
  ];

  return (
    <DashboardContainer>
      {/* Key Metrics */}
      <Row gutter={[24, 24]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card className="dashboard-card" loading={loading}>
            <Statistic
              title="Portfolio Value"
              value={15678.90}
              precision={2}
              valueStyle={{ color: '#52c41a' }}
              prefix={<DollarOutlined />}
              suffix={
                <span className="trending-up">
                  <ArrowUpOutlined /> 2.5%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="dashboard-card" loading={loading}>
            <Statistic
              title="KAIA Price"
              value={1.15}
              precision={3}
              valueStyle={{ color: '#667eea' }}
              prefix="$"
              suffix={
                <span className="trending-up">
                  <ArrowUpOutlined /> 3.2%
                </span>
              }
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="dashboard-card" loading={loading}>
            <Statistic
              title="Total Yield Earned"
              value={1234.56}
              precision={2}
              valueStyle={{ color: '#faad14' }}
              prefix={<TrendingUpOutlined />}
              suffix="KAIA"
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card className="dashboard-card" loading={loading}>
            <Statistic
              title="Active Positions"
              value={8}
              valueStyle={{ color: '#1890ff' }}
              prefix={<FireOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Charts Section */}
      <Row gutter={[24, 24]} style={{ marginBottom: 24 }}>
        <Col xs={24} lg={16}>
          <Card 
            title="KAIA Price Chart (24H)" 
            className="dashboard-card"
            loading={loading}
            extra={
              <Button type="primary" size="small">
                View Full Chart
              </Button>
            }
          >
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={priceData}>
                <defs>
                  <linearGradient id="colorPrice" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#667eea" stopOpacity={0.8}/>
                    <stop offset="95%" stopColor="#667eea" stopOpacity={0.1}/>
                  </linearGradient>
                </defs>
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
                />
                <Area 
                  type="monotone" 
                  dataKey="price" 
                  stroke="#667eea" 
                  fillOpacity={1} 
                  fill="url(#colorPrice)" 
                />
              </AreaChart>
            </ResponsiveContainer>
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card 
            title="Portfolio Distribution" 
            className="dashboard-card"
            loading={loading}
          >
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={portfolioData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={100}
                  paddingAngle={5}
                  dataKey="value"
                >
                  {portfolioData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip 
                  contentStyle={{
                    backgroundColor: 'rgba(26, 26, 46, 0.9)',
                    border: '1px solid rgba(255, 255, 255, 0.1)',
                    borderRadius: '8px',
                    color: '#ffffff'
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
            <div style={{ marginTop: 16 }}>
              {portfolioData.map((item, index) => (
                <div key={index} style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span style={{ display: 'flex', alignItems: 'center' }}>
                    <div 
                      style={{ 
                        width: 12, 
                        height: 12, 
                        backgroundColor: item.color, 
                        borderRadius: '50%', 
                        marginRight: 8 
                      }} 
                    />
                    {item.name}
                  </span>
                  <span>{item.value}%</span>
                </div>
              ))}
            </div>
          </Card>
        </Col>
      </Row>

      {/* Yield Opportunities and Recent Trades */}
      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card 
            title="Top Yield Opportunities" 
            className="dashboard-card"
            loading={loading}
            extra={
              <Button type="link">
                View All
              </Button>
            }
          >
            {yieldData.map((item, index) => (
              <div key={index} style={{ marginBottom: 16, padding: 16, background: 'rgba(102, 126, 234, 0.05)', borderRadius: 8 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <h4 style={{ margin: 0, color: '#ffffff' }}>{item.protocol}</h4>
                  <Tag color={item.risk === 'Low' ? 'green' : item.risk === 'Medium' ? 'orange' : 'blue'}>
                    {item.risk} Risk
                  </Tag>
                </div>
                <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                  <span>APY: <strong style={{ color: '#52c41a' }}>{item.apy}%</strong></span>
                  <span>TVL: <strong>${(item.tvl / 1000000).toFixed(1)}M</strong></span>
                </div>
                <Progress 
                  percent={item.apy * 5} 
                  strokeColor="#667eea" 
                  trailColor="rgba(255,255,255,0.1)"
                  showInfo={false}
                />
              </div>
            ))}
          </Card>
        </Col>
        <Col xs={24} lg={12}>
          <Card 
            title="Recent Trades" 
            className="dashboard-card"
            loading={loading}
            extra={
              <Button type="link">
                View All
              </Button>
            }
          >
            <Table
              columns={tradeColumns}
              dataSource={recentTrades}
              pagination={false}
              size="small"
              style={{ background: 'transparent' }}
            />
          </Card>
        </Col>
      </Row>
    </DashboardContainer>
  );
};

export default Dashboard;