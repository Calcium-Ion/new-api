import React, { useContext, useEffect, useState } from 'react';
import { Card, Col, Row } from '@douyinfe/semi-ui';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';

const Home = () => {
  const [statusState] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');

  const displayNotice = async () => {
    const res = await API.get('/api/notice');
    const { success, message, data } = res.data;
    if (success) {
      let oldNotice = localStorage.getItem('notice');
      if (data !== oldNotice && data !== '') {
        const htmlNotice = marked(data);
        showNotice(htmlNotice, true);
        localStorage.setItem('notice', data);
      }
    } else {
      showError(message);
    }
  };

  const displayHomePageContent = async () => {
    setHomePageContent(localStorage.getItem('home_page_content') || '');
    const res = await API.get('/api/home_page_content');
    const { success, message, data } = res.data;
    if (success) {
      let content = data;
      if (!data.startsWith('https://')) {
        content = marked.parse(data);
      }
      setHomePageContent(content);
      localStorage.setItem('home_page_content', content);
    } else {
      showError(message);
      setHomePageContent('加载首页内容失败...');
    }
    setHomePageContentLoaded(true);
  };

  const getStartTimeString = () => {
    const timestamp = statusState?.status?.start_time;
    return statusState.status ? timestamp2string(timestamp) : '';
  };

  useEffect(() => {
    displayNotice().then();
    displayHomePageContent().then();
  }, []);
  return (
    <>
      {
        homePageContentLoaded && homePageContent === '' ?
          <>
            <Card
              bordered={false}
              headerLine={false}
              title='系统状况'
              bodyStyle={{ padding: '10px 20px' }}
            >
              <Row gutter={16}>
                <Col span={12}>
                  <Card
                    title='系统信息'
                    headerExtraContent={<span
                      style={{ fontSize: '12px', color: 'var(--semi-color-text-1)' }}>系统信息总览</span>}>
                    <p>名称：{statusState?.status?.system_name}</p>
                    <p>版本：{statusState?.status?.version ? statusState?.status?.version : 'unknown'}</p>
                    <p>
                      源码：
                      <a
                        href='https://github.com/songquanpeng/one-api'
                        target='_blank' rel='noreferrer'
                      >
                        https://github.com/songquanpeng/one-api
                      </a>
                    </p>
                    <p>启动时间：{getStartTimeString()}</p>
                  </Card>
                </Col>
                <Col span={12}>
                  <Card
                    title='系统配置'
                    headerExtraContent={<span
                      style={{ fontSize: '12px', color: 'var(--semi-color-text-1)' }}>系统配置总览</span>}>
                    <p>
                      邮箱验证：
                      {statusState?.status?.email_verification === true ? '已启用' : '未启用'}
                    </p>
                    <p>
                      GitHub 身份验证：
                      {statusState?.status?.github_oauth === true ? '已启用' : '未启用'}
                    </p>
                    <p>
                      微信身份验证：
                      {statusState?.status?.wechat_login === true ? '已启用' : '未启用'}
                    </p>
                    <p>
                      Turnstile 用户校验：
                      {statusState?.status?.turnstile_check === true ? '已启用' : '未启用'}
                    </p>
                    <p>
                      Telegram 身份验证：
                      {statusState?.status?.telegram_oauth === true
                        ? '已启用' : '未启用'}
                    </p>
                  </Card>
                </Col>
              </Row>
            </Card>

          </>
          : <>
            {
              homePageContent.startsWith('https://') ?
                <iframe src={homePageContent} style={{ width: '100%', height: '100vh', border: 'none' }} /> :
                <div style={{ fontSize: 'larger' }} dangerouslySetInnerHTML={{ __html: homePageContent }}></div>
            }
          </>
      }

    </>
  );
};

export default Home;