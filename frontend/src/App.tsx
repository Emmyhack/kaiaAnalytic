import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { ConfigProvider, theme } from 'antd';
import styled from 'styled-components';

import Layout from './components/Layout/Layout';
import Dashboard from './pages/Dashboard/Dashboard';
import Analytics from './pages/Analytics/Analytics';
import Trading from './pages/Trading/Trading';
import Chat from './pages/Chat/Chat';
import Governance from './pages/Governance/Governance';
import Profile from './pages/Profile/Profile';

const AppContainer = styled.div`
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
`;

const App: React.FC = () => {
  return (
    <ConfigProvider
      theme={{
        algorithm: theme.darkAlgorithm,
        token: {
          colorPrimary: '#667eea',
          colorBgContainer: '#1a1a2e',
          colorBgLayout: '#16213e',
          colorText: '#ffffff',
          colorTextSecondary: '#a0a0a0',
        },
      }}
    >
      <AppContainer>
        <Router>
          <Layout>
            <Routes>
              <Route path="/" element={<Dashboard />} />
              <Route path="/analytics" element={<Analytics />} />
              <Route path="/trading" element={<Trading />} />
              <Route path="/chat" element={<Chat />} />
              <Route path="/governance" element={<Governance />} />
              <Route path="/profile" element={<Profile />} />
            </Routes>
          </Layout>
        </Router>
      </AppContainer>
    </ConfigProvider>
  );
};

export default App;