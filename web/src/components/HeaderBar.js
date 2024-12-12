import React, { useContext, useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { useSetTheme, useTheme } from '../context/Theme';

import { API, getLogo, getSystemName, isMobile, showSuccess } from '../helpers';
import '../index.css';

import fireworks from 'react-fireworks';

import {
  IconClose,
  IconHelpCircle,
  IconHome,
  IconHomeStroked, IconIndentLeft,
  IconKey, IconMenu,
  IconNoteMoneyStroked,
  IconPriceTag,
  IconUser
} from '@douyinfe/semi-icons';
import { Avatar, Button, Dropdown, Layout, Nav, Switch } from '@douyinfe/semi-ui';
import { stringToColor } from '../helpers/render';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { StyleContext } from '../context/Style/index.js';

// HeaderBar Buttons
let headerButtons = [
  {
    text: '关于',
    itemKey: 'about',
    to: '/about',
    icon: <IconHelpCircle />,
  },
];

if (localStorage.getItem('chat_link')) {
  headerButtons.splice(1, 0, {
    name: '聊天',
    to: '/chat',
    icon: 'comments',
  });
}

const HeaderBar = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  let navigate = useNavigate();

  const systemName = getSystemName();
  const logo = getLogo();
  const currentDate = new Date();
  // enable fireworks on new year(1.1 and 2.9-2.24)
  const isNewYear =
    (currentDate.getMonth() === 0 && currentDate.getDate() === 1) ||
    (currentDate.getMonth() === 1 &&
      currentDate.getDate() >= 9 &&
      currentDate.getDate() <= 24);

  let buttons = [
    {
      text: '首页',
      itemKey: 'home',
      to: '/',
    },
    {
      text: '控制台',
      itemKey: 'detail',
      to: '/',
    },
    {
      text: '定价',
      itemKey: 'pricing',
      to: '/pricing',
    },
  ];

  async function logout() {
    await API.get('/api/user/logout');
    showSuccess('注销成功!');
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
    }

    if (isNewYear) {
      console.log('Happy New Year!');
    }
  }, []);

  return (
    <>
      <Layout>
        <div style={{ width: '100%' }}>
          <Nav
            mode={'horizontal'}
            renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
              const routerMap = {
                about: '/about',
                login: '/login',
                register: '/register',
                pricing: '/pricing',
                detail: '/detail',
                home: '/',
              };
              return (
                <div onClick={(e) => {
                  if (props.itemKey === 'home') {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
                    styleDispatch({ type: 'SET_SIDER', payload: false });
                  } else {
                    styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
                    styleDispatch({ type: 'SET_SIDER', payload: true });
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
                <>
                  {
                    !styleState.showSider ?
                      <Button icon={<IconMenu />} theme="light" aria-label="展开侧边栏" onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: true })
                      } />:
                      <Button icon={<IconIndentLeft />} theme="light" aria-label="关闭侧边栏" onClick={
                        () => styleDispatch({ type: 'SET_SIDER', payload: false })
                      } />
                  }
                </>
              ),
            }:{
              logo: (
                <img src={logo} alt='logo' />
              ),
              text: systemName,
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
                    <Nav.Item itemKey={'new-year'} text={'🏮'} />
                  </Dropdown>
                )}
                <Nav.Item itemKey={'about'} icon={<IconHelpCircle />} />
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
                {userState.user ? (
                  <>
                    <Dropdown
                      position='bottomRight'
                      render={
                        <Dropdown.Menu>
                          <Dropdown.Item onClick={logout}>退出</Dropdown.Item>
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
                      text={'登录'}
                      // icon={<IconKey />}
                    />
                    <Nav.Item
                      itemKey={'register'}
                      text={'注册'}
                      icon={<IconUser />}
                    />
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
