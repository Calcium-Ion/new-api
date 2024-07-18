import React, { useEffect, useState, useRef } from 'react';
import {
  API,
  showError,
  showSuccess,
} from '../helpers';

import {
  Button,
} from '@douyinfe/semi-ui';

const TokensTable = () => {
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const loadAttemptsRef = useRef(0);
  const openLinkAttemptsRef = useRef(0);

  const loadTokens = async () => {
    if (loadAttemptsRef.current >= 2) return; // 最多尝试两次
    loadAttemptsRef.current++;
    setLoading(true);
    try {
      const res = await API.get('/api/token/');
      const { success, message, data } = res.data;
      if (success) {
        setTokens(data);
        if (data.length === 0) {
          showSuccess('初始化令牌成功！');
        } else {
          attemptOpenLink(data[0].key);
        }
      } else {
        showError(message);
        if (loadAttemptsRef.current < 2) {
          setTimeout(loadTokens, 1000); // 1秒后重试
        }
      }
    } catch (error) {
      showError('加载令牌失败: ' + error.message);
      if (loadAttemptsRef.current < 2) {
        setTimeout(loadTokens, 1000); // 1秒后重试
      }
    } finally {
      setLoading(false);
    }
  };

  const attemptOpenLink = (key) => {
    if (openLinkAttemptsRef.current >= 2) return; // 最多尝试两次
    openLinkAttemptsRef.current++;
    onOpenLink('next', key);
  };

  const onOpenLink = async (type, key) => {
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    const chatLink = localStorage.getItem('chat_link');
    const mjLink = localStorage.getItem('chat_link2');
    let defaultUrl;

    if (chatLink) {
      defaultUrl =
        chatLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
    }
    let url;
    switch (type) {
      case 'ama':
        url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
        break;
      case 'opencat':
        url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
        break;
      case 'lobe':
        url = `https://chat-preview.lobehub.com/?settings={"keyVaults":{"openai":{"apiKey":"sk-${key}","baseURL":"${encodedServerAddress}"}}}`;
        break;
      case 'next-mj':
        url =
          mjLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
        break;
      default:
        if (!chatLink) {
          showError('管理员未设置聊天链接');
          return;
        }
        url = defaultUrl;
    }

    try {
      window.open(url, '_blank');
    } catch (error) {
      showError('打开链接失败: ' + error.message);
      if (openLinkAttemptsRef.current < 2) {
        setTimeout(() => attemptOpenLink(key), 1000); // 1秒后重试
      }
    }
  };

  useEffect(() => {
    loadTokens();
  }, []);

  const handleButtonClick = () => {
    if (tokens.length > 0) {
      attemptOpenLink(tokens[0].key);
    } else {
      showError('没有可用的令牌进行对话。');
    }
  };

  return (
    <Button
      theme='light'
      type='primary'
      onClick={handleButtonClick}
      loading={loading}
      disabled={loading}
    >
      {loading ? '加载中...' : '开始对话'}
    </Button>
  );
};

export default TokensTable;
