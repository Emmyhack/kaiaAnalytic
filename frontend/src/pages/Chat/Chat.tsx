import React, { useState, useEffect, useRef } from 'react';
import { Card, Input, Button, List, Avatar, Tag, Spin, Space, Tooltip } from 'antd';
import {
  SendOutlined,
  RobotOutlined,
  UserOutlined,
  CopyOutlined,
  ThunderboltOutlined,
  QuestionCircleOutlined,
} from '@ant-design/icons';
import styled from 'styled-components';
import { io, Socket } from 'socket.io-client';

const { TextArea } = Input;

const ChatContainer = styled.div`
  height: calc(100vh - 200px);
  display: flex;
  flex-direction: column;
`;

const ChatMessages = styled.div`
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: rgba(26, 26, 46, 0.3);
  border-radius: 12px;
  margin-bottom: 16px;
  
  &::-webkit-scrollbar {
    width: 6px;
  }
  
  &::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 3px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: rgba(102, 126, 234, 0.6);
    border-radius: 3px;
  }
`;

const MessageBubble = styled.div<{ isUser: boolean }>`
  max-width: 70%;
  margin: 16px 0;
  padding: 12px 16px;
  border-radius: 18px;
  background: ${props => props.isUser 
    ? 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
    : 'rgba(26, 26, 46, 0.8)'
  };
  border: 1px solid ${props => props.isUser 
    ? 'rgba(102, 126, 234, 0.3)'
    : 'rgba(255, 255, 255, 0.1)'
  };
  align-self: ${props => props.isUser ? 'flex-end' : 'flex-start'};
  margin-left: ${props => props.isUser ? 'auto' : '0'};
  margin-right: ${props => props.isUser ? '0' : 'auto'};
  position: relative;
  
  &::before {
    content: '';
    position: absolute;
    ${props => props.isUser ? 'right: -8px' : 'left: -8px'};
    top: 50%;
    transform: translateY(-50%);
    border: 8px solid transparent;
    border-${props => props.isUser ? 'left' : 'right'}-color: ${props => props.isUser 
      ? '#667eea'
      : 'rgba(26, 26, 46, 0.8)'
    };
  }
`;

const ChatInput = styled.div`
  display: flex;
  gap: 12px;
  align-items: flex-end;
`;

const SuggestionChips = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 16px 0;
`;

const SuggestionChip = styled(Tag)`
  cursor: pointer;
  padding: 4px 12px;
  border-radius: 16px;
  background: rgba(102, 126, 234, 0.1);
  border: 1px solid rgba(102, 126, 234, 0.3);
  color: #ffffff;
  
  &:hover {
    background: rgba(102, 126, 234, 0.2);
    border-color: rgba(102, 126, 234, 0.5);
  }
`;

const ActionButtons = styled.div`
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 12px;
`;

interface ChatMessage {
  id: string;
  message: string;
  response?: string;
  isUser: boolean;
  timestamp: Date;
  intent?: string;
  entities?: any[];
  actions?: any[];
  suggestions?: string[];
  loading?: boolean;
}

const Chat: React.FC = () => {
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      id: '1',
      message: 'Welcome to KaiaAnalyticsAI! I can help you with yield farming opportunities, trading suggestions, governance information, and execute on-chain actions. What would you like to know?',
      isUser: false,
      timestamp: new Date(),
      suggestions: [
        'Show me yield opportunities',
        'Get trading suggestions for KAIA',
        'Check governance proposals',
        'What\'s the KAIA price?'
      ]
    }
  ]);
  const [inputValue, setInputValue] = useState('');
  const [loading, setLoading] = useState(false);
  const [socket, setSocket] = useState<Socket | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    // Initialize WebSocket connection
    const newSocket = io('ws://localhost:8080', {
      query: { user_id: '0x1234567890123456789012345678901234567890' }
    });

    newSocket.on('connect', () => {
      console.log('Connected to chat server');
    });

    newSocket.on('message', (data: any) => {
      const assistantMessage: ChatMessage = {
        id: Date.now().toString(),
        message: data.response,
        isUser: false,
        timestamp: new Date(),
        intent: data.intent,
        entities: data.entities,
        actions: data.actions,
        suggestions: data.suggestions,
      };

      setMessages(prev => {
        const updated = [...prev];
        const lastMessage = updated[updated.length - 1];
        if (lastMessage.loading) {
          updated[updated.length - 1] = assistantMessage;
        } else {
          updated.push(assistantMessage);
        }
        return updated;
      });
      setLoading(false);
    });

    setSocket(newSocket);

    return () => {
      newSocket.close();
    };
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  const handleSendMessage = async () => {
    if (!inputValue.trim() || loading) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      message: inputValue,
      isUser: true,
      timestamp: new Date(),
    };

    const loadingMessage: ChatMessage = {
      id: (Date.now() + 1).toString(),
      message: '',
      isUser: false,
      timestamp: new Date(),
      loading: true,
    };

    setMessages(prev => [...prev, userMessage, loadingMessage]);
    setLoading(true);
    setInputValue('');

    try {
      if (socket) {
        socket.emit('message', { message: inputValue });
      } else {
        // Fallback to HTTP API
        const response = await fetch('/api/v1/chat/query', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            message: inputValue,
            user_id: '0x1234567890123456789012345678901234567890',
          }),
        });

        const data = await response.json();
        
        const assistantMessage: ChatMessage = {
          id: Date.now().toString(),
          message: data.response,
          isUser: false,
          timestamp: new Date(),
          intent: data.intent,
          entities: data.entities,
          actions: data.actions,
          suggestions: data.suggestions,
        };

        setMessages(prev => {
          const updated = [...prev];
          updated[updated.length - 1] = assistantMessage;
          return updated;
        });
        setLoading(false);
      }
    } catch (error) {
      console.error('Error sending message:', error);
      setMessages(prev => {
        const updated = [...prev];
        updated[updated.length - 1] = {
          id: Date.now().toString(),
          message: 'Sorry, I encountered an error. Please try again.',
          isUser: false,
          timestamp: new Date(),
        };
        return updated;
      });
      setLoading(false);
    }
  };

  const handleSuggestionClick = (suggestion: string) => {
    setInputValue(suggestion);
  };

  const handleActionClick = async (action: any) => {
    try {
      const response = await fetch('/api/v1/chat/action', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          type: action.type,
          parameters: action.parameters,
          user_id: '0x1234567890123456789012345678901234567890',
        }),
      });

      const data = await response.json();
      
      const actionResultMessage: ChatMessage = {
        id: Date.now().toString(),
        message: `Action "${action.type}" ${data.status}: ${data.result}`,
        isUser: false,
        timestamp: new Date(),
      };

      setMessages(prev => [...prev, actionResultMessage]);
    } catch (error) {
      console.error('Error executing action:', error);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  return (
    <ChatContainer>
      <ChatMessages>
        {messages.map((message) => (
          <div key={message.id} style={{ display: 'flex', flexDirection: 'column' }}>
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: 8 }}>
              <Avatar
                icon={message.isUser ? <UserOutlined /> : <RobotOutlined />}
                style={{
                  backgroundColor: message.isUser ? '#667eea' : '#52c41a',
                  marginRight: 8,
                }}
              />
              <span style={{ fontSize: 12, color: '#a0a0a0' }}>
                {message.timestamp.toLocaleTimeString()}
              </span>
              {!message.isUser && (
                <Tooltip title="Copy message">
                  <Button
                    type="text"
                    size="small"
                    icon={<CopyOutlined />}
                    onClick={() => copyToClipboard(message.message)}
                    style={{ marginLeft: 8, color: '#a0a0a0' }}
                  />
                </Tooltip>
              )}
            </div>
            
            <MessageBubble isUser={message.isUser}>
              {message.loading ? (
                <Space>
                  <Spin size="small" />
                  <span>AI is thinking...</span>
                </Space>
              ) : (
                <div>
                  <div style={{ whiteSpace: 'pre-wrap' }}>{message.message}</div>
                  
                  {message.actions && message.actions.length > 0 && (
                    <ActionButtons>
                      {message.actions.map((action, index) => (
                        <Button
                          key={index}
                          type="primary"
                          size="small"
                          icon={<ThunderboltOutlined />}
                          onClick={() => handleActionClick(action)}
                          style={{
                            background: 'rgba(102, 126, 234, 0.2)',
                            border: '1px solid rgba(102, 126, 234, 0.3)',
                          }}
                        >
                          {action.type.replace('_', ' ')}
                        </Button>
                      ))}
                    </ActionButtons>
                  )}
                </div>
              )}
            </MessageBubble>

            {message.suggestions && message.suggestions.length > 0 && (
              <SuggestionChips>
                {message.suggestions.map((suggestion, index) => (
                  <SuggestionChip
                    key={index}
                    onClick={() => handleSuggestionClick(suggestion)}
                  >
                    {suggestion}
                  </SuggestionChip>
                ))}
              </SuggestionChips>
            )}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </ChatMessages>

      <ChatInput>
        <TextArea
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          onPressEnter={(e) => {
            if (!e.shiftKey) {
              e.preventDefault();
              handleSendMessage();
            }
          }}
          placeholder="Ask me about yield opportunities, trading signals, governance, or request on-chain actions..."
          autoSize={{ minRows: 1, maxRows: 4 }}
          style={{
            background: 'rgba(26, 26, 46, 0.8)',
            border: '1px solid rgba(255, 255, 255, 0.1)',
            color: '#ffffff',
          }}
        />
        
        <Button
          type="primary"
          icon={<SendOutlined />}
          onClick={handleSendMessage}
          loading={loading}
          disabled={!inputValue.trim()}
          style={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            border: 'none',
            height: 'auto',
            minHeight: 32,
          }}
        >
          Send
        </Button>
      </ChatInput>

      <div style={{ marginTop: 16, padding: 16, background: 'rgba(26, 26, 46, 0.3)', borderRadius: 8 }}>
        <Space>
          <QuestionCircleOutlined style={{ color: '#a0a0a0' }} />
          <span style={{ fontSize: 12, color: '#a0a0a0' }}>
            Try asking: "What are the best yield opportunities?" or "Show me KAIA trading signals"
          </span>
        </Space>
      </div>
    </ChatContainer>
  );
};

export default Chat;