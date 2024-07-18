import React, { useEffect, useState } from 'react';
import {
  API,
  showError,
  showSuccess,
} from '../helpers';

import {
  Button,
  Modal,
} from '@douyinfe/semi-ui';

const TokensTable = () => {
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [hasOpenedLink, setHasOpenedLink] = useState(false);

  const loadTokens = async () => {
    setLoading(true);
    const res = await API.get('/api/token/');
    const { success, message, data } = res.data;
    if (success) {
      setTokens(data);
    } else {
      showError(message);
    }
    setLoading(false);
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

    window.open(url, '_blank');
  };

  useEffect(() => {
    loadTokens()
      .then(() => {
        if (tokens.length > 0 && !hasOpenedLink) {
          onOpenLink('next', tokens[0].key);
          setHasOpenedLink(true); // 设置状态以避免重复调用
        } else if (tokens.length === 0) {
          showError('没有可用的令牌进行对话。');
        }
      })
      .catch((reason) => {
        showError(reason);
      });
  }, [tokens, hasOpenedLink]);

  return (
    <>
      <Button
        theme='light'
        type='primary'
        onClick={() => {
          if (tokens.length > 0) {
            onOpenLink('next', tokens[0].key);
          } else {
            showError('没有可用的令牌进行对话。');
          }
        }}
      >
        开始聊天
      </Button>
    </>
  );
};

export default TokensTable;
