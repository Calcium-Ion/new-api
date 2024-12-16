import React, { useContext, useEffect, useState } from 'react';
import { Card, Col, Row } from '@douyinfe/semi-ui';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';
import { StyleContext } from '../../context/Style/index.js';
import { useTranslation } from 'react-i18next';

const Home = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [homePageContentLoaded, setHomePageContentLoaded] = useState(false);
  const [homePageContent, setHomePageContent] = useState('');
  const [styleState, styleDispatch] = useContext(StyleContext);

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

        // 如果内容是 URL，则发送主题模式
        if (data.startsWith('https://')) {
            const iframe = document.querySelector('iframe');
            if (iframe) {
                const theme = localStorage.getItem('theme-mode') || 'light';
                // 测试是否正确传递theme-mode给iframe
                // console.log('Sending theme-mode to iframe:', theme); 
                iframe.onload = () => {
                    iframe.contentWindow.postMessage({ themeMode: theme }, '*');
                    iframe.contentWindow.postMessage({ lang: i18n.language }, '*');
                };
            }
        }
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
  });

  return (
    <>
      {homePageContentLoaded && homePageContent === '' ? (
        <>
          <Card
            bordered={false}
            headerLine={false}
            title={t('系统状况')}
            bodyStyle={{ padding: '10px 20px' }}
          >
            <Row gutter={16}>
              <Col span={12}>
                <Card
                  title={t('系统信息')}
                  headerExtraContent={
                    <span
                      style={{
                        fontSize: '12px',
                        color: 'var(--semi-color-text-1)',
                      }}
                    >
                      {t('系统信息总览')}
                    </span>
                  }
                >
                  <p>{t('名称')}：{statusState?.status?.system_name}</p>
                  <p>
                    {t('版本')}：
                    {statusState?.status?.version
                      ? statusState?.status?.version
                      : 'unknown'}
                  </p>
                  <p>
                    {t('源码')}：
                    <a
                      href='https://github.com/Calcium-Ion/new-api'
                      target='_blank'
                      rel='noreferrer'
                    >
                      https://github.com/Calcium-Ion/new-api
                    </a>
                  </p>
                  <p>
                    {t('协议')}：
                    <a
                      href='https://www.apache.org/licenses/LICENSE-2.0'
                      target='_blank'
                      rel='noreferrer'
                    >
                      Apache-2.0 License
                    </a>
                  </p>
                  <p>{t('启动时间')}：{getStartTimeString()}</p>
                </Card>
              </Col>
              <Col span={12}>
                <Card
                  title={t('系统配置')}
                  headerExtraContent={
                    <span
                      style={{
                        fontSize: '12px',
                        color: 'var(--semi-color-text-1)',
                      }}
                    >
                      {t('系统配置总览')}
                    </span>
                  }
                >
                  <p>
                    {t('邮箱验证')}：
                    {statusState?.status?.email_verification === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                  <p>
                    {t('GitHub 身份验证')}：
                    {statusState?.status?.github_oauth === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                  <p>
                    {t('微信身份验证')}：
                    {statusState?.status?.wechat_login === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                  <p>
                    {t('Turnstile 用户校验')}：
                    {statusState?.status?.turnstile_check === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                  <p>
                    {t('Telegram 身份验证')}：
                    {statusState?.status?.telegram_oauth === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                  <p>
                    {t('Linux DO 身份验证')}：
                    {statusState?.status?.linuxdo_oauth === true
                      ? t('已启用')
                      : t('未启用')}
                  </p>
                </Card>
              </Col>
            </Row>
          </Card>
        </>
      ) : (
        <>
          {homePageContent.startsWith('https://') ? (
            <iframe
              src={homePageContent}
              style={{ width: '100%', height: '100vh', border: 'none' }}
            />
          ) : (
            <div
              style={{ fontSize: 'larger' }}
              dangerouslySetInnerHTML={{ __html: homePageContent }}
            ></div>
          )}
        </>
      )}
    </>
  );
};

export default Home;
