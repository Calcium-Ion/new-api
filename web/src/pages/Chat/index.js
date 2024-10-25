import React, {useEffect} from 'react';
import { useTokenKeys } from '../../components/fetchTokenKeys';
import {Banner, Layout} from '@douyinfe/semi-ui';
import { useParams } from 'react-router-dom';

const ChatPage = () => {
  const { id } = useParams();
  const { keys, serverAddress, isLoading } = useTokenKeys(id);

  const comLink = (key) => {
    // console.log('chatLink:', chatLink);
    if (!serverAddress || !key) return '';
      let link = localStorage.getItem('chat_link');
      if (link) {
          link = `${link}/#/?settings={"key":"sk-${key}","url":"${encodeURIComponent(serverAddress)}"}`;
      } else if (id) {
          let chats = localStorage.getItem('chats');
          if (chats) {
              chats = JSON.parse(chats);
              if (Array.isArray(chats) && chats.length > 0) {
                  for (let k in chats[id]) {
                      link = chats[id][k];
                      link = link.replace('{address}', encodeURIComponent(serverAddress));
                      link = link.replace('{key}', 'sk-' + key);
                  }
              }
          }
      }
      return link;
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
          <Banner
              description={"正在跳转......"}
              type={"warning"}
          />
        </Layout.Header>
      </Layout>
    </div>
  );
};

export default ChatPage;