import React, { useState } from 'react';
import { Row, Col, Card, Button, Table, Tag, Statistic, Progress, List, Avatar } from 'antd';
import {
  TrendingUpOutlined,
  TrendingDownOutlined,
  ThunderboltOutlined,
  StarOutlined,
  BellOutlined,
} from '@ant-design/icons';
import styled from 'styled-components';

const TradingContainer = styled.div`
  .trading-card {
    background: rgba(26, 26, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    margin-bottom: 24px;
  }
  
  .signal-card {
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    border: 1px solid rgba(102, 126, 234, 0.3);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
  }
  
  .bull-signal {
    border-left: 4px solid #52c41a;
  }
  
  .bear-signal {
    border-left: 4px solid #ff4d4f;
  }
  
  .neutral-signal {
    border-left: 4px solid #faad14;
  }
`;

const tradingSignals = [
  {
    id: '1',
    pair: 'KAIA/USDC',
    signal: 'Strong Buy',
    confidence: 78,
    targetPrice: 1.25,
    stopLoss: 0.95,
    timeHorizon: '1-2 weeks',
    reason: 'Positive technical indicators, increasing volume',
    type: 'bull',
  },
  {
    id: '2',
    pair: 'ETH/KAIA',
    signal: 'Hold',
    confidence: 65,
    targetPrice: 1850,
    stopLoss: 1650,
    timeHorizon: '2-4 weeks',
    reason: 'Consolidation phase, awaiting breakout',
    type: 'neutral',
  },
  {
    id: '3',
    pair: 'BTC/KAIA',
    signal: 'Buy',
    confidence: 72,
    targetPrice: 28500,
    stopLoss: 25000,
    timeHorizon: '3-5 weeks',
    reason: 'Support level holding, bullish momentum',
    type: 'bull',
  },
];

const portfolioPerformance = [
  {
    key: '1',
    asset: 'KAIA',
    amount: '1,234.56',
    value: '$1,420.25',
    change24h: '+2.5%',
    allocation: 45,
    color: '#52c41a',
  },
  {
    key: '2',
    asset: 'USDC',
    amount: '850.00',
    value: '$850.00',
    change24h: '0.0%',
    allocation: 25,
    color: '#1890ff',
  },
  {
    key: '3',
    asset: 'ETH',
    amount: '0.45',
    value: '$765.30',
    change24h: '+1.8%',
    allocation: 20,
    color: '#722ed1',
  },
  {
    key: '4',
    asset: 'BTC',
    amount: '0.012',
    value: '$342.60',
    change24h: '+0.9%',
    allocation: 10,
    color: '#fa8c16',
  },
];

const recentAlerts = [
  {
    id: '1',
    message: 'KAIA price reached target of $1.20',
    type: 'success',
    time: '5 min ago',
  },
  {
    id: '2',
    message: 'High volume detected in KAIA/USDC pair',
    type: 'info',
    time: '15 min ago',
  },
  {
    id: '3',
    message: 'New arbitrage opportunity: KaiaSwap vs KaiaLend',
    type: 'warning',
    time: '1 hour ago',
  },
];

const Trading: React.FC = () => {
  const [selectedSignal, setSelectedSignal] = useState<string | null>(null);

  const portfolioColumns = [
    {
      title: 'Asset',
      dataIndex: 'asset',
      key: 'asset',
      render: (asset: string) => (
        <div style={{ display: 'flex', alignItems: 'center' }}>
          <Avatar style={{ backgroundColor: '#667eea', marginRight: 8 }}>
            {asset}
          </Avatar>
          {asset}
        </div>
      ),
    },
    {
      title: 'Amount',
      dataIndex: 'amount',
      key: 'amount',
    },
    {
      title: 'Value',
      dataIndex: 'value',
      key: 'value',
    },
    {
      title: '24h Change',
      dataIndex: 'change24h',
      key: 'change24h',
      render: (change: string) => (
        <span style={{ color: change.includes('+') ? '#52c41a' : change.includes('-') ? '#ff4d4f' : '#a0a0a0' }}>
          {change}
        </span>
      ),
    },
    {
      title: 'Allocation',
      dataIndex: 'allocation',
      key: 'allocation',
      render: (allocation: number, record: any) => (
        <div>
          <Progress
            percent={allocation}
            strokeColor={record.color}
            trailColor="rgba(255,255,255,0.1)"
            size="small"
            showInfo={false}
          />
          <span style={{ fontSize: '12px', marginTop: '4px', display: 'block' }}>
            {allocation}%
          </span>
        </div>
      ),
    },
  ];

  return (
    <TradingContainer>
      <Row gutter={[24, 24]}>
        {/* Portfolio Overview */}
        <Col xs={24} lg={8}>
          <Card title="Portfolio Overview" className="trading-card">
            <Row gutter={16}>
              <Col span={12}>
                <Statistic
                  title="Total Value"
                  value={3378.15}
                  precision={2}
                  prefix="$"
                  valueStyle={{ color: '#52c41a', fontSize: '24px' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="24h Change"
                  value={2.3}
                  precision={1}
                  suffix="%"
                  prefix={<TrendingUpOutlined />}
                  valueStyle={{ color: '#52c41a', fontSize: '24px' }}
                />
              </Col>
            </Row>
            <Table
              columns={portfolioColumns}
              dataSource={portfolioPerformance}
              pagination={false}
              size="small"
              style={{ marginTop: 16 }}
            />
          </Card>
        </Col>

        {/* Trading Signals */}
        <Col xs={24} lg={16}>
          <Card 
            title="AI Trading Signals" 
            className="trading-card"
            extra={
              <Button type="primary" icon={<ThunderboltOutlined />}>
                Generate New Signals
              </Button>
            }
          >
            {tradingSignals.map((signal) => (
              <div
                key={signal.id}
                className={`signal-card ${signal.type}-signal`}
                style={{ cursor: 'pointer' }}
                onClick={() => setSelectedSignal(signal.id)}
              >
                <Row justify="space-between" align="middle">
                  <Col span={18}>
                    <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
                      <h4 style={{ margin: 0, marginRight: 12 }}>{signal.pair}</h4>
                      <Tag 
                        color={signal.type === 'bull' ? 'green' : signal.type === 'bear' ? 'red' : 'orange'}
                        style={{ marginRight: 8 }}
                      >
                        {signal.signal}
                      </Tag>
                      <span style={{ fontSize: '12px', color: '#a0a0a0' }}>
                        {signal.confidence}% confidence
                      </span>
                    </div>
                    <div style={{ fontSize: '14px', color: '#a0a0a0', marginBottom: 8 }}>
                      Target: ${signal.targetPrice} | Stop Loss: ${signal.stopLoss} | {signal.timeHorizon}
                    </div>
                    <div style={{ fontSize: '12px', color: '#ffffff' }}>
                      {signal.reason}
                    </div>
                  </Col>
                  <Col span={6} style={{ textAlign: 'right' }}>
                    <Progress
                      type="circle"
                      percent={signal.confidence}
                      width={50}
                      strokeColor={signal.type === 'bull' ? '#52c41a' : signal.type === 'bear' ? '#ff4d4f' : '#faad14'}
                      trailColor="rgba(255,255,255,0.1)"
                    />
                  </Col>
                </Row>
              </div>
            ))}
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]} style={{ marginTop: 0 }}>
        {/* Recent Alerts */}
        <Col xs={24} lg={12}>
          <Card 
            title="Recent Alerts" 
            className="trading-card"
            extra={
              <Button type="link" icon={<BellOutlined />}>
                Manage Alerts
              </Button>
            }
          >
            <List
              dataSource={recentAlerts}
              renderItem={(alert) => (
                <List.Item style={{ padding: '8px 0', borderBottom: '1px solid rgba(255,255,255,0.1)' }}>
                  <List.Item.Meta
                    avatar={
                      <Avatar 
                        style={{ 
                          backgroundColor: alert.type === 'success' ? '#52c41a' : 
                                          alert.type === 'warning' ? '#faad14' : '#1890ff' 
                        }}
                        icon={<BellOutlined />}
                      />
                    }
                    title={
                      <span style={{ color: '#ffffff', fontSize: '14px' }}>
                        {alert.message}
                      </span>
                    }
                    description={
                      <span style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        {alert.time}
                      </span>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </Col>

        {/* Quick Actions */}
        <Col xs={24} lg={12}>
          <Card title="Quick Actions" className="trading-card">
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Button 
                  type="primary" 
                  size="large" 
                  block
                  style={{ 
                    background: 'linear-gradient(135deg, #52c41a 0%, #389e0d 100%)',
                    border: 'none',
                    height: '60px'
                  }}
                >
                  <TrendingUpOutlined style={{ fontSize: '20px' }} />
                  <br />
                  Buy KAIA
                </Button>
              </Col>
              <Col span={12}>
                <Button 
                  size="large" 
                  block
                  style={{ 
                    background: 'linear-gradient(135deg, #ff4d4f 0%, #cf1322 100%)',
                    border: 'none',
                    color: '#ffffff',
                    height: '60px'
                  }}
                >
                  <TrendingDownOutlined style={{ fontSize: '20px' }} />
                  <br />
                  Sell KAIA
                </Button>
              </Col>
              <Col span={12}>
                <Button 
                  size="large" 
                  block
                  style={{ 
                    background: 'rgba(102, 126, 234, 0.2)',
                    border: '1px solid rgba(102, 126, 234, 0.3)',
                    color: '#ffffff',
                    height: '60px'
                  }}
                >
                  <StarOutlined style={{ fontSize: '20px' }} />
                  <br />
                  Yield Farm
                </Button>
              </Col>
              <Col span={12}>
                <Button 
                  size="large" 
                  block
                  style={{ 
                    background: 'rgba(250, 173, 20, 0.2)',
                    border: '1px solid rgba(250, 173, 20, 0.3)',
                    color: '#ffffff',
                    height: '60px'
                  }}
                >
                  <ThunderboltOutlined style={{ fontSize: '20px' }} />
                  <br />
                  Auto Trade
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </TradingContainer>
  );
};

export default Trading;