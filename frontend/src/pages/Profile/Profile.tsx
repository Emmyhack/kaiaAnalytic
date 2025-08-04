import React, { useState } from 'react';
import { Row, Col, Card, Button, Switch, Input, Avatar, Badge, Tag, Tabs, Table, Modal } from 'antd';
import {
  UserOutlined,
  SettingOutlined,
  CrownOutlined,
  WalletOutlined,
  BellOutlined,
  SecurityScanOutlined,
  HistoryOutlined,
  GiftOutlined,
} from '@ant-design/icons';
import styled from 'styled-components';

const { TabPane } = Tabs;

const ProfileContainer = styled.div`
  .profile-card {
    background: rgba(26, 26, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    margin-bottom: 24px;
  }
  
  .subscription-card {
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    border: 2px solid rgba(102, 126, 234, 0.3);
    border-radius: 12px;
    padding: 24px;
    text-align: center;
    position: relative;
    
    &.premium {
      border-color: #faad14;
      background: linear-gradient(135deg, rgba(250, 173, 20, 0.1) 0%, rgba(255, 193, 7, 0.1) 100%);
    }
  }
  
  .feature-list {
    text-align: left;
    margin: 16px 0;
  }
  
  .feature-item {
    display: flex;
    align-items: center;
    margin-bottom: 8px;
    color: #a0a0a0;
  }
  
  .feature-item.included {
    color: #52c41a;
  }
`;

const subscriptionTiers = [
  {
    id: 'free',
    name: 'Free Tier',
    price: 0,
    period: 'Forever',
    features: [
      { name: 'Basic Analytics Dashboard', included: true },
      { name: 'Transaction Volume Charts', included: true },
      { name: 'Gas Trend Analysis', included: true },
      { name: 'Limited AI Chat (5 queries/day)', included: true },
      { name: 'Advanced Yield Analytics', included: false },
      { name: 'Trading Signals', included: false },
      { name: 'On-chain Actions', included: false },
      { name: 'Premium Support', included: false },
    ],
  },
  {
    id: 'premium',
    name: 'Premium Tier',
    price: 100,
    period: 'month',
    features: [
      { name: 'All Free Features', included: true },
      { name: 'Advanced Yield Analytics', included: true },
      { name: 'AI Trading Signals', included: true },
      { name: 'Unlimited AI Chat', included: true },
      { name: 'On-chain Actions', included: true },
      { name: 'Portfolio Optimization', included: true },
      { name: 'Real-time Alerts', included: true },
      { name: 'Premium Support', included: true },
    ],
  },
];

const transactionHistory = [
  {
    key: '1',
    date: '2024-01-15',
    type: 'Subscription',
    description: 'Premium Tier Renewal',
    amount: '-100 KAIA',
    status: 'Completed',
  },
  {
    key: '2',
    date: '2024-01-10',
    type: 'Action',
    description: 'Staking via AI Chat',
    amount: '-500 KAIA',
    status: 'Completed',
  },
  {
    key: '3',
    date: '2024-01-05',
    type: 'Reward',
    description: 'Yield Farming Rewards',
    amount: '+25.5 KAIA',
    status: 'Completed',
  },
];

const Profile: React.FC = () => {
  const [currentTier, setCurrentTier] = useState('premium');
  const [upgradeModalVisible, setUpgradeModalVisible] = useState(false);
  const [notifications, setNotifications] = useState({
    trading: true,
    governance: true,
    yield: false,
    security: true,
  });

  const handleNotificationChange = (key: string, value: boolean) => {
    setNotifications(prev => ({ ...prev, [key]: value }));
  };

  const handleUpgrade = (tierId: string) => {
    setUpgradeModalVisible(true);
  };

  const historyColumns = [
    {
      title: 'Date',
      dataIndex: 'date',
      key: 'date',
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={type === 'Subscription' ? 'blue' : type === 'Action' ? 'purple' : 'green'}>
          {type}
        </Tag>
      ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
    },
    {
      title: 'Amount',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: string) => (
        <span style={{ color: amount.startsWith('+') ? '#52c41a' : '#ff4d4f' }}>
          {amount}
        </span>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => (
        <Tag color="green">{status}</Tag>
      ),
    },
  ];

  return (
    <ProfileContainer>
      <Row gutter={[24, 24]}>
        {/* Profile Overview */}
        <Col xs={24} lg={8}>
          <Card className="profile-card">
            <div style={{ textAlign: 'center', marginBottom: 24 }}>
              <Avatar
                size={80}
                icon={<UserOutlined />}
                style={{
                  backgroundColor: '#667eea',
                  marginBottom: 16,
                }}
              />
              <h3 style={{ color: '#ffffff', margin: 0 }}>0x1234...5678</h3>
              <div style={{ color: '#a0a0a0', marginTop: 4 }}>
                Member since January 2024
              </div>
              <Badge
                count={currentTier === 'premium' ? 'PREMIUM' : 'FREE'}
                style={{
                  backgroundColor: currentTier === 'premium' ? '#faad14' : '#1890ff',
                  marginTop: 8,
                }}
              />
            </div>

            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span style={{ color: '#a0a0a0' }}>Wallet Balance</span>
                <span style={{ color: '#52c41a', fontWeight: 'bold' }}>1,234.56 KAIA</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span style={{ color: '#a0a0a0' }}>Queries Used</span>
                <span style={{ color: '#ffffff' }}>
                  {currentTier === 'premium' ? '∞' : '3/5'} this month
                </span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <span style={{ color: '#a0a0a0' }}>Actions Executed</span>
                <span style={{ color: '#ffffff' }}>12 this month</span>
              </div>
            </div>

            <Button
              type="primary"
              block
              icon={<SettingOutlined />}
              style={{
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                border: 'none',
              }}
            >
              Account Settings
            </Button>
          </Card>
        </Col>

        {/* Main Content */}
        <Col xs={24} lg={16}>
          <Tabs defaultActiveKey="subscription" size="large">
            <TabPane tab="Subscription" key="subscription">
              <Row gutter={[24, 24]}>
                {subscriptionTiers.map((tier) => (
                  <Col xs={24} md={12} key={tier.id}>
                    <div className={`subscription-card ${tier.id === 'premium' ? 'premium' : ''}`}>
                      {tier.id === 'premium' && (
                        <div style={{ position: 'absolute', top: -12, right: 20 }}>
                          <Tag color="#faad14" icon={<CrownOutlined />}>
                            PREMIUM
                          </Tag>
                        </div>
                      )}
                      
                      <h3 style={{ color: '#ffffff', marginBottom: 8 }}>{tier.name}</h3>
                      <div style={{ fontSize: '32px', fontWeight: 'bold', color: tier.id === 'premium' ? '#faad14' : '#667eea', marginBottom: 4 }}>
                        {tier.price === 0 ? 'Free' : `${tier.price} KAIA`}
                      </div>
                      <div style={{ color: '#a0a0a0', marginBottom: 20 }}>
                        {tier.price === 0 ? 'Forever' : `per ${tier.period}`}
                      </div>

                      <div className="feature-list">
                        {tier.features.map((feature, index) => (
                          <div
                            key={index}
                            className={`feature-item ${feature.included ? 'included' : ''}`}
                          >
                            <span style={{ marginRight: 8 }}>
                              {feature.included ? '✓' : '✗'}
                            </span>
                            {feature.name}
                          </div>
                        ))}
                      </div>

                      {currentTier === tier.id ? (
                        <Button
                          type="default"
                          block
                          disabled
                          style={{
                            background: 'rgba(82, 196, 26, 0.2)',
                            border: '1px solid #52c41a',
                            color: '#52c41a',
                          }}
                        >
                          Current Plan
                        </Button>
                      ) : (
                        <Button
                          type="primary"
                          block
                          onClick={() => handleUpgrade(tier.id)}
                          style={{
                            background: tier.id === 'premium' 
                              ? 'linear-gradient(135deg, #faad14 0%, #ffc107 100%)'
                              : 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                            border: 'none',
                          }}
                        >
                          {tier.id === 'premium' ? 'Upgrade to Premium' : 'Downgrade to Free'}
                        </Button>
                      )}
                    </div>
                  </Col>
                ))}
              </Row>
            </TabPane>

            <TabPane tab="Notifications" key="notifications">
              <Card title="Notification Preferences" className="profile-card">
                <div style={{ marginBottom: 24 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Trading Alerts</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Get notified about trading opportunities and market movements
                      </div>
                    </div>
                    <Switch
                      checked={notifications.trading}
                      onChange={(value) => handleNotificationChange('trading', value)}
                    />
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Governance Updates</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Notifications about new proposals and voting deadlines
                      </div>
                    </div>
                    <Switch
                      checked={notifications.governance}
                      onChange={(value) => handleNotificationChange('governance', value)}
                    />
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Yield Opportunities</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Alerts about high-yield farming opportunities
                      </div>
                    </div>
                    <Switch
                      checked={notifications.yield}
                      onChange={(value) => handleNotificationChange('yield', value)}
                    />
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Security Alerts</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Important security notifications and account activity
                      </div>
                    </div>
                    <Switch
                      checked={notifications.security}
                      onChange={(value) => handleNotificationChange('security', value)}
                    />
                  </div>
                </div>

                <Button
                  type="primary"
                  icon={<BellOutlined />}
                  style={{
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    border: 'none',
                  }}
                >
                  Save Preferences
                </Button>
              </Card>
            </TabPane>

            <TabPane tab="Transaction History" key="history">
              <Card 
                title="Transaction History" 
                className="profile-card"
                extra={
                  <Button type="link" icon={<HistoryOutlined />}>
                    Export History
                  </Button>
                }
              >
                <Table
                  columns={historyColumns}
                  dataSource={transactionHistory}
                  pagination={{ pageSize: 10 }}
                />
              </Card>
            </TabPane>

            <TabPane tab="Security" key="security">
              <Card title="Security Settings" className="profile-card">
                <div style={{ marginBottom: 24 }}>
                  <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
                    <SecurityScanOutlined style={{ color: '#52c41a', fontSize: '20px', marginRight: 12 }} />
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Wallet Connected</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Your wallet is securely connected via Kaikas
                      </div>
                    </div>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
                    <WalletOutlined style={{ color: '#667eea', fontSize: '20px', marginRight: 12 }} />
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>Two-Factor Authentication</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Add an extra layer of security to your account
                      </div>
                    </div>
                    <Button
                      type="primary"
                      size="small"
                      style={{ marginLeft: 'auto' }}
                    >
                      Enable 2FA
                    </Button>
                  </div>

                  <div style={{ display: 'flex', alignItems: 'center', marginBottom: 16 }}>
                    <GiftOutlined style={{ color: '#faad14', fontSize: '20px', marginRight: 12 }} />
                    <div>
                      <div style={{ color: '#ffffff', fontWeight: 'bold' }}>API Keys</div>
                      <div style={{ color: '#a0a0a0', fontSize: '12px' }}>
                        Manage API keys for external integrations
                      </div>
                    </div>
                    <Button
                      type="default"
                      size="small"
                      style={{ marginLeft: 'auto' }}
                    >
                      Manage Keys
                    </Button>
                  </div>
                </div>

                <Button
                  type="primary"
                  icon={<SecurityScanOutlined />}
                  style={{
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    border: 'none',
                  }}
                >
                  Security Audit
                </Button>
              </Card>
            </TabPane>
          </Tabs>
        </Col>
      </Row>

      {/* Upgrade Modal */}
      <Modal
        title="Upgrade Subscription"
        open={upgradeModalVisible}
        onOk={() => setUpgradeModalVisible(false)}
        onCancel={() => setUpgradeModalVisible(false)}
        okText="Confirm Upgrade"
        okButtonProps={{
          style: {
            background: 'linear-gradient(135deg, #faad14 0%, #ffc107 100%)',
            border: 'none',
          }
        }}
      >
        <div style={{ marginBottom: 24 }}>
          <div style={{ fontSize: '16px', marginBottom: 12 }}>
            Upgrade to Premium Tier for 100 KAIA/month
          </div>
          <div style={{ color: '#a0a0a0', marginBottom: 16 }}>
            You'll get access to all premium features including unlimited AI chat, 
            advanced analytics, and on-chain actions.
          </div>
          <div style={{ background: 'rgba(102, 126, 234, 0.1)', padding: 16, borderRadius: 8 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <span>Monthly Subscription</span>
              <span>100 KAIA</span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
              <span>Your Balance</span>
              <span>1,234.56 KAIA</span>
            </div>
            <div style={{ borderTop: '1px solid rgba(255,255,255,0.1)', paddingTop: 8, marginTop: 8 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontWeight: 'bold' }}>
                <span>Remaining Balance</span>
                <span>1,134.56 KAIA</span>
              </div>
            </div>
          </div>
        </div>
      </Modal>
    </ProfileContainer>
  );
};

export default Profile;