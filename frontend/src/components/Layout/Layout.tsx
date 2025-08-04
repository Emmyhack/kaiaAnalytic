import React, { useState } from 'react';
import { Layout as AntLayout, Menu, Button, Avatar, Dropdown, Badge } from 'antd';
import {
  DashboardOutlined,
  BarChartOutlined,
  TradingViewOutlined,
  MessageOutlined,
  TeamOutlined,
  UserOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  BellOutlined,
  WalletOutlined,
} from '@ant-design/icons';
import { useNavigate, useLocation } from 'react-router-dom';
import styled from 'styled-components';

const { Header, Sider, Content } = AntLayout;

const StyledLayout = styled(AntLayout)`
  min-height: 100vh;
  background: transparent;
`;

const StyledSider = styled(Sider)`
  background: rgba(26, 26, 46, 0.9) !important;
  backdrop-filter: blur(20px);
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  
  .ant-layout-sider-trigger {
    background: rgba(102, 126, 234, 0.2);
    color: #ffffff;
    border-top: 1px solid rgba(255, 255, 255, 0.1);
  }
`;

const StyledHeader = styled(Header)`
  background: rgba(26, 26, 46, 0.9) !important;
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
`;

const Logo = styled.div`
  color: #667eea;
  font-size: 24px;
  font-weight: bold;
  margin: 16px;
  text-align: center;
  
  .logo-text {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
    background-clip: text;
  }
`;

const HeaderActions = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
`;

const WalletInfo = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  background: rgba(102, 126, 234, 0.1);
  border-radius: 8px;
  border: 1px solid rgba(102, 126, 234, 0.3);
`;

const StyledContent = styled(Content)`
  margin: 24px;
  padding: 24px;
  background: rgba(26, 26, 46, 0.3);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: auto;
`;

interface LayoutProps {
  children: React.ReactNode;
}

const Layout: React.FC<LayoutProps> = ({ children }) => {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();

  const menuItems = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      key: '/analytics',
      icon: <BarChartOutlined />,
      label: 'Analytics',
    },
    {
      key: '/trading',
      icon: <TradingViewOutlined />,
      label: 'Trading',
    },
    {
      key: '/chat',
      icon: <MessageOutlined />,
      label: 'AI Chat',
    },
    {
      key: '/governance',
      icon: <TeamOutlined />,
      label: 'Governance',
    },
    {
      key: '/profile',
      icon: <UserOutlined />,
      label: 'Profile',
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key);
  };

  const userMenuItems = [
    {
      key: 'profile',
      label: 'Profile Settings',
      icon: <UserOutlined />,
    },
    {
      key: 'wallet',
      label: 'Wallet Settings',
      icon: <WalletOutlined />,
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      label: 'Logout',
      danger: true,
    },
  ];

  const handleUserMenuClick = ({ key }: { key: string }) => {
    if (key === 'logout') {
      // Handle logout
      console.log('Logout clicked');
    } else if (key === 'profile') {
      navigate('/profile');
    }
  };

  return (
    <StyledLayout>
      <StyledSider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={250}
        collapsedWidth={80}
      >
        <Logo>
          <div className="logo-text">
            {collapsed ? 'KA' : 'KaiaAnalyticsAI'}
          </div>
        </Logo>
        
        <Menu
          theme="dark"
          mode="inline"
          selectedKeys={[location.pathname]}
          items={menuItems}
          onClick={handleMenuClick}
          style={{
            background: 'transparent',
            border: 'none',
          }}
        />
      </StyledSider>

      <AntLayout>
        <StyledHeader>
          <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{
                fontSize: '16px',
                width: 64,
                height: 64,
                color: '#ffffff',
              }}
            />
            
            <h2 style={{ color: '#ffffff', margin: 0 }}>
              {menuItems.find(item => item.key === location.pathname)?.label || 'Dashboard'}
            </h2>
          </div>

          <HeaderActions>
            <WalletInfo>
              <WalletOutlined />
              <span>1,234.56 KAIA</span>
            </WalletInfo>

            <Badge count={3} size="small">
              <Button
                type="text"
                icon={<BellOutlined />}
                style={{ color: '#ffffff' }}
              />
            </Badge>

            <Dropdown
              menu={{
                items: userMenuItems,
                onClick: handleUserMenuClick,
              }}
              placement="bottomRight"
            >
              <Avatar
                style={{
                  backgroundColor: '#667eea',
                  cursor: 'pointer',
                }}
                icon={<UserOutlined />}
              />
            </Dropdown>
          </HeaderActions>
        </StyledHeader>

        <StyledContent>
          {children}
        </StyledContent>
      </AntLayout>
    </StyledLayout>
  );
};

export default Layout;