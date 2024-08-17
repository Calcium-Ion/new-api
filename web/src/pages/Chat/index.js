import React from 'react';
import { useTokenKeys } from '../../components/fetchTokenKeys';
import {  Layout } from '@douyinfe/semi-ui';

const ChatPage = () => {
  const { keys, chatLink, serverAddress, isLoading } = useTokenKeys();

  const comLink = (key) => {
    if (!chatLink || !serverAddress || !key) return '';
    return `${chatLink}/#/?settings={"key":"sk-${key}","url":"${encodeURIComponent(serverAddress)}"}`;
  };

  const iframeSrc = keys.length > 0 ? comLink(keys[0]) : '';

  return !isLoading && iframeSrc ? (
    <iframe
      src={iframeSrc}
      style={{ width: '100%', height: '85vh', border: 'none' }}
      title="Token Frame"
    />
  ) : (
    <div>
      <Layout>
        <Layout.Header>
          <h3 style={{ color: 'red'}}>
            当前没有可用的已启用令牌，请确认是否有令牌处于启用状态！<br />
            正在跳转......
          </h3>
        </Layout.Header>
      </Layout>
    </div>
  );
};

export default ChatPage;