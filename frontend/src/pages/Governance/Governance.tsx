import React, { useState } from 'react';
import { Row, Col, Card, Button, Table, Tag, Progress, Avatar, Modal, Radio, Input } from 'antd';
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  VoteOutlined,
  UserOutlined,
  CalendarOutlined,
} from '@ant-design/icons';
import styled from 'styled-components';

const { TextArea } = Input;

const GovernanceContainer = styled.div`
  .governance-card {
    background: rgba(26, 26, 46, 0.8);
    border: 1px solid rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    margin-bottom: 24px;
  }
  
  .proposal-card {
    background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
    border: 1px solid rgba(102, 126, 234, 0.3);
    border-radius: 12px;
    padding: 20px;
    margin-bottom: 16px;
    cursor: pointer;
    transition: all 0.3s ease;
    
    &:hover {
      border-color: rgba(102, 126, 234, 0.5);
      transform: translateY(-2px);
    }
  }
  
  .vote-option {
    padding: 12px;
    border-radius: 8px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    margin-bottom: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    
    &:hover {
      background: rgba(102, 126, 234, 0.1);
    }
    
    &.selected {
      border-color: #667eea;
      background: rgba(102, 126, 234, 0.2);
    }
  }
`;

const proposals = [
  {
    id: 'KIP-001',
    title: 'Increase Block Gas Limit',
    description: 'Proposal to increase the block gas limit from 30M to 40M to accommodate higher transaction throughput and reduce congestion during peak usage.',
    category: 'Technical',
    author: '0x1234...5678',
    created: '2024-01-01',
    endDate: '2024-01-15',
    status: 'Active',
    votesFor: 67500000,
    votesAgainst: 15200000,
    totalVotes: 82700000,
    quorum: 100000000,
    sentiment: 0.82,
    participation: 0.68,
    details: 'This proposal aims to improve network performance by increasing the gas limit, allowing for more complex transactions and better DeFi interactions.',
  },
  {
    id: 'KIP-002',
    title: 'Fee Structure Update',
    description: 'Modify transaction fee structure to implement a dynamic pricing model based on network congestion.',
    category: 'Economic',
    author: '0xabcd...efgh',
    created: '2023-12-20',
    endDate: '2024-01-10',
    status: 'Passed',
    votesFor: 89200000,
    votesAgainst: 8500000,
    totalVotes: 97700000,
    quorum: 100000000,
    sentiment: 0.91,
    participation: 0.71,
    details: 'Implementation of EIP-1559 style fee structure to make transaction costs more predictable and fair.',
  },
  {
    id: 'KIP-003',
    title: 'Validator Rewards Distribution',
    description: 'Adjust validator reward distribution to better incentivize network security and decentralization.',
    category: 'Governance',
    author: '0x9876...5432',
    created: '2024-01-05',
    endDate: '2024-01-20',
    status: 'Under Review',
    votesFor: 45600000,
    votesAgainst: 23400000,
    totalVotes: 69000000,
    quorum: 100000000,
    sentiment: 0.66,
    participation: 0.59,
    details: 'Proposing changes to reward distribution to encourage smaller validators and improve network decentralization.',
  },
];

const Governance: React.FC = () => {
  const [selectedProposal, setSelectedProposal] = useState<any>(null);
  const [voteModalVisible, setVoteModalVisible] = useState(false);
  const [selectedVote, setSelectedVote] = useState<string>('');
  const [voteReason, setVoteReason] = useState('');

  const handleProposalClick = (proposal: any) => {
    setSelectedProposal(proposal);
  };

  const handleVoteClick = (proposal: any) => {
    setSelectedProposal(proposal);
    setVoteModalVisible(true);
  };

  const handleVoteSubmit = async () => {
    if (!selectedVote) return;
    
    // Here you would call the smart contract to submit the vote
    console.log('Submitting vote:', {
      proposalId: selectedProposal.id,
      vote: selectedVote,
      reason: voteReason,
    });
    
    setVoteModalVisible(false);
    setSelectedVote('');
    setVoteReason('');
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'Active': return 'green';
      case 'Passed': return 'blue';
      case 'Failed': return 'red';
      case 'Under Review': return 'orange';
      default: return 'default';
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'Technical': return 'blue';
      case 'Economic': return 'green';
      case 'Governance': return 'purple';
      default: return 'default';
    }
  };

  return (
    <GovernanceContainer>
      <Row gutter={[24, 24]}>
        {/* Governance Overview */}
        <Col span={24}>
          <Card title="Governance Overview" className="governance-card">
            <Row gutter={[24, 24]}>
              <Col xs={24} sm={6}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', fontWeight: 'bold', color: '#667eea' }}>
                    {proposals.length}
                  </div>
                  <div style={{ color: '#a0a0a0' }}>Active Proposals</div>
                </div>
              </Col>
              <Col xs={24} sm={6}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', fontWeight: 'bold', color: '#52c41a' }}>
                    1,234
                  </div>
                  <div style={{ color: '#a0a0a0' }}>Total Voters</div>
                </div>
              </Col>
              <Col xs={24} sm={6}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', fontWeight: 'bold', color: '#faad14' }}>
                    68.5%
                  </div>
                  <div style={{ color: '#a0a0a0' }}>Avg Participation</div>
                </div>
              </Col>
              <Col xs={24} sm={6}>
                <div style={{ textAlign: 'center' }}>
                  <div style={{ fontSize: '32px', fontWeight: 'bold', color: '#1890ff' }}>
                    156M
                  </div>
                  <div style={{ color: '#a0a0a0' }}>KAIA Voting Power</div>
                </div>
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      <Row gutter={[24, 24]}>
        {/* Proposals List */}
        <Col xs={24} lg={selectedProposal ? 14 : 24}>
          <Card title="Active Proposals" className="governance-card">
            {proposals.map((proposal) => (
              <div
                key={proposal.id}
                className="proposal-card"
                onClick={() => handleProposalClick(proposal)}
              >
                <Row justify="space-between" align="top">
                  <Col span={18}>
                    <div style={{ display: 'flex', alignItems: 'center', marginBottom: 12 }}>
                      <h3 style={{ margin: 0, marginRight: 12, color: '#ffffff' }}>
                        {proposal.id}: {proposal.title}
                      </h3>
                      <Tag color={getStatusColor(proposal.status)} style={{ marginRight: 8 }}>
                        {proposal.status}
                      </Tag>
                      <Tag color={getCategoryColor(proposal.category)}>
                        {proposal.category}
                      </Tag>
                    </div>
                    
                    <div style={{ color: '#a0a0a0', marginBottom: 16, lineHeight: '1.5' }}>
                      {proposal.description}
                    </div>
                    
                    <Row gutter={16} style={{ marginBottom: 16 }}>
                      <Col span={8}>
                        <div style={{ fontSize: '12px', color: '#a0a0a0' }}>Author</div>
                        <div style={{ display: 'flex', alignItems: 'center' }}>
                          <Avatar size="small" icon={<UserOutlined />} style={{ marginRight: 8 }} />
                          {proposal.author}
                        </div>
                      </Col>
                      <Col span={8}>
                        <div style={{ fontSize: '12px', color: '#a0a0a0' }}>Created</div>
                        <div style={{ display: 'flex', alignItems: 'center' }}>
                          <CalendarOutlined style={{ marginRight: 8 }} />
                          {proposal.created}
                        </div>
                      </Col>
                      <Col span={8}>
                        <div style={{ fontSize: '12px', color: '#a0a0a0' }}>Ends</div>
                        <div style={{ color: new Date(proposal.endDate) > new Date() ? '#52c41a' : '#ff4d4f' }}>
                          {proposal.endDate}
                        </div>
                      </Col>
                    </Row>
                    
                    <div style={{ marginBottom: 12 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                        <span style={{ fontSize: '12px', color: '#a0a0a0' }}>Voting Progress</span>
                        <span style={{ fontSize: '12px', color: '#a0a0a0' }}>
                          {((proposal.totalVotes / proposal.quorum) * 100).toFixed(1)}% of quorum
                        </span>
                      </div>
                      <Progress
                        percent={(proposal.totalVotes / proposal.quorum) * 100}
                        strokeColor="#667eea"
                        trailColor="rgba(255,255,255,0.1)"
                        showInfo={false}
                      />
                    </div>
                    
                    <Row gutter={16}>
                      <Col span={12}>
                        <div style={{ fontSize: '12px', color: '#52c41a', marginBottom: 4 }}>
                          For: {((proposal.votesFor / proposal.totalVotes) * 100).toFixed(1)}%
                        </div>
                        <Progress
                          percent={(proposal.votesFor / proposal.totalVotes) * 100}
                          strokeColor="#52c41a"
                          trailColor="rgba(255,255,255,0.1)"
                          showInfo={false}
                          size="small"
                        />
                      </Col>
                      <Col span={12}>
                        <div style={{ fontSize: '12px', color: '#ff4d4f', marginBottom: 4 }}>
                          Against: {((proposal.votesAgainst / proposal.totalVotes) * 100).toFixed(1)}%
                        </div>
                        <Progress
                          percent={(proposal.votesAgainst / proposal.totalVotes) * 100}
                          strokeColor="#ff4d4f"
                          trailColor="rgba(255,255,255,0.1)"
                          showInfo={false}
                          size="small"
                        />
                      </Col>
                    </Row>
                  </Col>
                  
                  <Col span={6} style={{ textAlign: 'right' }}>
                    {proposal.status === 'Active' && (
                      <Button
                        type="primary"
                        icon={<VoteOutlined />}
                        onClick={(e) => {
                          e.stopPropagation();
                          handleVoteClick(proposal);
                        }}
                        style={{
                          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                          border: 'none',
                        }}
                      >
                        Vote
                      </Button>
                    )}
                  </Col>
                </Row>
              </div>
            ))}
          </Card>
        </Col>

        {/* Proposal Details */}
        {selectedProposal && (
          <Col xs={24} lg={10}>
            <Card 
              title={`${selectedProposal.id} Details`} 
              className="governance-card"
              extra={
                <Button 
                  type="text" 
                  onClick={() => setSelectedProposal(null)}
                  style={{ color: '#a0a0a0' }}
                >
                  ×
                </Button>
              }
            >
              <div style={{ marginBottom: 24 }}>
                <h3 style={{ color: '#ffffff', marginBottom: 8 }}>{selectedProposal.title}</h3>
                <div style={{ color: '#a0a0a0', lineHeight: '1.6' }}>
                  {selectedProposal.details}
                </div>
              </div>

              <div style={{ marginBottom: 24 }}>
                <h4 style={{ color: '#ffffff', marginBottom: 12 }}>Voting Results</h4>
                <div style={{ marginBottom: 12 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span style={{ color: '#52c41a' }}>For</span>
                    <span>{(selectedProposal.votesFor / 1000000).toFixed(1)}M KAIA</span>
                  </div>
                  <Progress
                    percent={(selectedProposal.votesFor / selectedProposal.totalVotes) * 100}
                    strokeColor="#52c41a"
                    trailColor="rgba(255,255,255,0.1)"
                    showInfo={false}
                  />
                </div>
                <div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                    <span style={{ color: '#ff4d4f' }}>Against</span>
                    <span>{(selectedProposal.votesAgainst / 1000000).toFixed(1)}M KAIA</span>
                  </div>
                  <Progress
                    percent={(selectedProposal.votesAgainst / selectedProposal.totalVotes) * 100}
                    strokeColor="#ff4d4f"
                    trailColor="rgba(255,255,255,0.1)"
                    showInfo={false}
                  />
                </div>
              </div>

              <div style={{ marginBottom: 24 }}>
                <h4 style={{ color: '#ffffff', marginBottom: 12 }}>Community Sentiment</h4>
                <Progress
                  percent={selectedProposal.sentiment * 100}
                  strokeColor={selectedProposal.sentiment > 0.7 ? '#52c41a' : selectedProposal.sentiment > 0.5 ? '#faad14' : '#ff4d4f'}
                  trailColor="rgba(255,255,255,0.1)"
                  format={(percent) => `${percent?.toFixed(0)}% Positive`}
                />
              </div>

              {selectedProposal.status === 'Active' && (
                <Button
                  type="primary"
                  icon={<VoteOutlined />}
                  block
                  size="large"
                  onClick={() => handleVoteClick(selectedProposal)}
                  style={{
                    background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                    border: 'none',
                  }}
                >
                  Cast Your Vote
                </Button>
              )}
            </Card>
          </Col>
        )}
      </Row>

      {/* Vote Modal */}
      <Modal
        title={`Vote on ${selectedProposal?.id}`}
        open={voteModalVisible}
        onOk={handleVoteSubmit}
        onCancel={() => setVoteModalVisible(false)}
        okText="Submit Vote"
        cancelText="Cancel"
        okButtonProps={{
          disabled: !selectedVote,
          style: {
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
          }
        }}
      >
        <div style={{ marginBottom: 24 }}>
          <h4 style={{ marginBottom: 16 }}>Select your vote:</h4>
          
          <div
            className={`vote-option ${selectedVote === 'for' ? 'selected' : ''}`}
            onClick={() => setSelectedVote('for')}
          >
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <CheckCircleOutlined style={{ color: '#52c41a', fontSize: '20px', marginRight: 12 }} />
              <div>
                <div style={{ fontWeight: 'bold', color: '#52c41a' }}>Vote For</div>
                <div style={{ fontSize: '12px', color: '#a0a0a0' }}>
                  Support this proposal
                </div>
              </div>
            </div>
          </div>

          <div
            className={`vote-option ${selectedVote === 'against' ? 'selected' : ''}`}
            onClick={() => setSelectedVote('against')}
          >
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: '20px', marginRight: 12 }} />
              <div>
                <div style={{ fontWeight: 'bold', color: '#ff4d4f' }}>Vote Against</div>
                <div style={{ fontSize: '12px', color: '#a0a0a0' }}>
                  Oppose this proposal
                </div>
              </div>
            </div>
          </div>

          <div
            className={`vote-option ${selectedVote === 'abstain' ? 'selected' : ''}`}
            onClick={() => setSelectedVote('abstain')}
          >
            <div style={{ display: 'flex', alignItems: 'center' }}>
              <ExclamationCircleOutlined style={{ color: '#faad14', fontSize: '20px', marginRight: 12 }} />
              <div>
                <div style={{ fontWeight: 'bold', color: '#faad14' }}>Abstain</div>
                <div style={{ fontSize: '12px', color: '#a0a0a0' }}>
                  Participate without taking a side
                </div>
              </div>
            </div>
          </div>
        </div>

        <div>
          <h4 style={{ marginBottom: 8 }}>Reason (optional):</h4>
          <TextArea
            rows={4}
            value={voteReason}
            onChange={(e) => setVoteReason(e.target.value)}
            placeholder="Explain your voting decision..."
            style={{
              background: 'rgba(26, 26, 46, 0.8)',
              border: '1px solid rgba(255, 255, 255, 0.1)',
              color: '#ffffff',
            }}
          />
        </div>
      </Modal>
    </GovernanceContainer>
  );
};

export default Governance;