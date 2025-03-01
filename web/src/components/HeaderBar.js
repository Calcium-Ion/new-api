import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { useSetTheme, useTheme } from '../context/Theme';
import { useTranslation } from 'react-i18next';

import { API, getLogo, getSystemName, isMobile, showSuccess } from '../helpers';
import '../index.css';

import fireworks from 'react-fireworks';

import {
  IconClose,
  IconHelpCircle,
  IconHome,
  IconHomeStroked, IconIndentLeft,
  IconComment,
  IconKey, IconMenu,
  IconNoteMoneyStroked,
  IconPriceTag,
  IconUser,
  IconLanguage
} from '@douyinfe/semi-icons';
import { Avatar, Button, Dropdown, Layout, Nav, Switch, Tag } from '@douyinfe/semi-ui';
import { stringToColor } from '../helpers/render';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { StyleContext } from '../context/Style/index.js';
import { StatusContext } from '../context/Status/index.js';

const HeaderBar = () => {
  const { t, i18n } = useTranslation();
  const [userState, userDispatch] = useContext(UserContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  let navigate = useNavigate();
  const [currentLang, setCurrentLang] = useState(i18n.language);

  const systemName = getSystemName();
  const logo = getLogo();
  const currentDate = new Date();
  // enable fireworks on new year(1.1 and 2.9-2.24)
  const isNewYear =
    (currentDate.getMonth() === 0 && currentDate.getDate() === 1);

  // Check if self-use mode is enabled
  const isSelfUseMode = statusState?.status?.self_use_mode_enabled || false;
  const isDemoSiteMode = statusState?.status?.demo_site_enabled || false;

  let buttons = [
    {
      text: t('首页'),
      itemKey: 'home',
      to: '/',
    },
    {
      text: t('控制台'),
      itemKey: 'detail',
      to: '/',
    },
    {
      text: t('定价'),
      itemKey: 'pricing',
      to: '/pricing',
    },
    {
      text: t('关于'),
      itemKey: 'about',
      to: '/about',
    },
  ];

  async function logout() {
    await API.get('/api/user/logout');
    showSuccess(t('注销成功!'));
    userDispatch({ type: 'logout' });
    localStorage.removeItem('user');
    navigate('/login');
  }

  const handleNewYearClick = () => {
    fireworks.init('root', {});
    fireworks.start();
    setTimeout(() => {
      fireworks.stop();
      setTimeout(() => {
        window.location.reload();
      }, 10000);
    }, 3000);
  };

  const theme = useTheme();
  const setTheme = useSetTheme();

  useEffect(() => {
    if (theme === 'dark') {
      document.body.setAttribute('theme-mode', 'dark');
    } else {
      document.body.removeAttribute('theme-mode');
    }
    // 发送当前主题模式给子页面
    const iframe = document.querySelector('iframe');
    if (iframe) {
      iframe.contentWindow.postMessage({ themeMode: theme }, '*');
    }

    if (isNewYear) {
      console.log('Happy New Year!');
    }
  }, [theme]);

  useEffect(() => {
    const handleLanguageChanged = (lng) => {
      setCurrentLang(lng);
      const iframe = document.querySelector('iframe');
      if (iframe) {
        iframe.contentWindow.postMessage({ lang: lng }, '*');
      }
    };

    i18n.on('languageChanged', handleLanguageChanged);

    return () => {
      i18n.off('languageChanged', handleLanguageChanged);
    };
  }, [i18n]);

  const handleLanguageChange = (lang) => {
    i18n.changeLanguage(lang);
  };

  return (
    <>
      <Layout>
        <div style={{ width: '100%' }}>
          <Nav
            className={'topnav'}
            mode={'horizontal'}
            renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
              const routerMap = {
                about: '/about',
                login: '/login',
                register: '/register',
                pricing: '/pricing',
                detail: '/detail',
                home: '/',
                chat: '/chat',
              };
              return (
                <div onClick={(e) => {
                  if (props.itemKey === 'home') {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
                    styleDispatch({ type: 'SET_SIDER', payload: false });
                  } else {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
                    if (!styleState.isMobile) {
                      styleDispatch({ type: 'SET_SIDER', payload: true });
                    }
                  }
                }}>
                  <Link
                    className="header-bar-text"
                    style={{ textDecoration: 'none' }}
                    to={routerMap[props.itemKey]}
                  >
                    {itemElement}
                  </Link>
                </div>
              );
            }}
            selectedKeys={[]}
            // items={headerButtons}
            onSelect={(key) => {}}
            header={styleState.isMobile?{
              logo: (
                <div style={{ display: 'flex', alignItems: 'center', position: 'relative' }}>
                  {
                    !styleState.showSider ?
                      <Button icon={<IconMenu />} theme="light" aria-label={t('展开侧边栏')} onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: true })
                      } />:
                      <Button icon={<IconIndentLeft />} theme="light" aria-label={t('闭侧边栏')} onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: false })
                      } />
                  }
                  {(isSelfUseMode || isDemoSiteMode) && (
                    <Tag 
                      color={isSelfUseMode ? 'purple' : 'blue'}
                      style={{ 
                        position: 'absolute',
                        top: '-8px',
                        right: '-15px',
                        fontSize: '0.7rem',
                        padding: '0 4px',
                        height: 'auto',
                        lineHeight: '1.2',
                        zIndex: 1,
                        pointerEvents: 'none'
                      }}
                    >
                      {isSelfUseMode ? t('自用模式') : t('演示站点')}
                    </Tag>
                  )}
                </div>
              ),
            }:{
              logo: (
                <img src={logo} alt='logo' />
              ),
              text: (
                <div style={{ position: 'relative', display: 'inline-block' }}>
                  {systemName}
                  {(isSelfUseMode || isDemoSiteMode) && (
                    <Tag 
                      color={isSelfUseMode ? 'purple' : 'blue'}
                      style={{ 
                        position: 'absolute', 
                        top: '-10px', 
                        right: '-25px', 
                        fontSize: '0.7rem',
                        padding: '0 4px',
                        whiteSpace: 'nowrap',
                        zIndex: 1,
                        boxShadow: '0 0 3px rgba(255, 255, 255, 0.7)'
                      }}
                    >
                      {isSelfUseMode ? t('自用模式') : t('演示站点')}
                    </Tag>
                  )}
                </div>
              ),
            }}
            items={buttons}
            footer={
              <>
                {isNewYear && (
                  // happy new year
                  <Dropdown
                    position='bottomRight'
                    render={
                      <Dropdown.Menu>
                        <Dropdown.Item onClick={handleNewYearClick}>
                          Happy New Year!!!
                        </Dropdown.Item>
                      </Dropdown.Menu>
                    }
                  >
                    <Nav.Item itemKey={'new-year'} text={'🎉'} />
                  </Dropdown>
                )}
                {/* <Nav.Item itemKey={'about'} icon={<IconHelpCircle />} /> */}
                <>
                  <Switch
                    checkedText='🌞'
                    size={styleState.isMobile?'default':'large'}
                    checked={theme === 'dark'}
                    uncheckedText='🌙'
                    onChange={(checked) => {
                      setTheme(checked);
                    }}
                  />
                </>
                <Dropdown
                  position='bottomRight'
                  render={
                    <Dropdown.Menu>
                      <Dropdown.Item
                        onClick={() => handleLanguageChange('zh')}
                        type={currentLang === 'zh' ? 'primary' : 'tertiary'}
                      >
                        中文
                      </Dropdown.Item>
                      <Dropdown.Item
                        onClick={() => handleLanguageChange('en')}
                        type={currentLang === 'en' ? 'primary' : 'tertiary'}
                      >
                        English
                      </Dropdown.Item>
                    </Dropdown.Menu>
                  }
                >
                  <Nav.Item
                    itemKey={'language'}
                    icon={<IconLanguage />}
                  />
                </Dropdown>
                {userState.user ? (
                  <>
                    <Dropdown
                      position='bottomRight'
                      render={
                        <Dropdown.Menu>
                          <Dropdown.Item onClick={logout}>{t('退出')}</Dropdown.Item>
                        </Dropdown.Menu>
                      }
                    >
                      <Avatar
                        size='small'
                        color={stringToColor(userState.user.username)}
                        style={{ margin: 4 }}
                      >
                        {userState.user.username[0]}
                      </Avatar>
                      {styleState.isMobile?null:<Text>{userState.user.username}</Text>}
                    </Dropdown>
                  </>
                ) : (
                  <>
                    <Nav.Item
                      itemKey={'login'}
                      text={!styleState.isMobile?t('登录'):null}
                      icon={<IconUser />}
                    />
                    {
                      // Hide register option in self-use mode
                      !styleState.isMobile && !isSelfUseMode && (
                        <Nav.Item
                          itemKey={'register'}
                          text={t('注册')}
                          icon={<IconKey />}
                        />
                      )
                    }
                  </>
                )}
              </>
            }
          ></Nav>
        </div>
      </Layout>
    </>
  );
};

export default HeaderBar;
