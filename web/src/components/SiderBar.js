import React, {useContext, useState} from 'react';
import {Link, useNavigate} from 'react-router-dom';
import {UserContext} from '../context/User';

import {Button, Container, Icon, Menu, Segment} from 'semantic-ui-react';
import {API, getLogo, getSystemName, isAdmin, isMobile, showSuccess} from '../helpers';
import '../index.css';

import {
    IconAt,
    IconHistogram,
    IconGift,
    IconKey,
    IconUser,
    IconLayers,
    IconSetting,
    IconCreditCard,
    IconSemiLogo,
    IconHome,
    IconImage
} from '@douyinfe/semi-icons';
import {Nav, Avatar, Dropdown, Layout} from '@douyinfe/semi-ui';

// HeaderBar Buttons
let headerButtons = [
    {
        text: '首页',
        itemKey: 'home',
        to: '/',
        icon: <IconHome/>
    },
    {
        text: '渠道',
        itemKey: 'channel',
        to: '/channel',
        icon: <IconLayers/>,
        admin: true
    },

    {
        text: '令牌',
        itemKey: 'token',
        to: '/token',
        icon: <IconKey/>
    },
    {
        text: '兑换码',
        itemKey: 'redemption',
        to: '/redemption',
        icon: <IconGift/>,
        admin: true
    },
    {
        text: '钱包',
        itemKey: 'topup',
        to: '/topup',
        icon: <IconCreditCard/>
    },
    {
        text: '用户管理',
        itemKey: 'user',
        to: '/user',
        icon: <IconUser/>,
        admin: true
    },
    {
        text: '日志',
        itemKey: 'log',
        to: '/log',
        icon: <IconHistogram/>
    },
    {
        text: '绘图',
        itemKey: 'midjourney',
        to: '/midjourney',
        icon: <IconImage/>
    },
    {
        text: '设置',
        itemKey: 'setting',
        to: '/setting',
        icon: <IconSetting/>
    },
    // {
    //     text: '关于',
    //     itemKey: 'about',
    //     to: '/about',
    //     icon: <IconAt/>
    // }
];

if (localStorage.getItem('chat_link')) {
    headerButtons.splice(1, 0, {
        name: '聊天',
        to: '/chat',
        icon: 'comments'
    });
}

const HeaderBar = () => {
    const [userState, userDispatch] = useContext(UserContext);
    let navigate = useNavigate();
    const [selectedKeys, setSelectedKeys] = useState(['home']);
    const [showSidebar, setShowSidebar] = useState(false);
    const systemName = getSystemName();
    const logo = getLogo();

    async function logout() {
        setShowSidebar(false);
        await API.get('/api/user/logout');
        showSuccess('注销成功!');
        userDispatch({type: 'logout'});
        localStorage.removeItem('user');
        navigate('/login');
    }

    const toggleSidebar = () => {
        setShowSidebar(!showSidebar);
    };

    const renderButtons = (isMobile) => {
        return headerButtons.map((button) => {
            if (button.admin && !isAdmin()) return <></>;
            if (isMobile) {
                return (
                    <Menu.Item
                        onClick={() => {
                            navigate(button.to);
                            setShowSidebar(false);
                        }}
                    >
                        {button.name}
                    </Menu.Item>
                );
            }
            return (
                <Menu.Item key={button.name} as={Link} to={button.to}>
                    <Icon name={button.icon}/>
                    {button.name}
                </Menu.Item>
            );
        });
    };

    if (isMobile()) {
        return (
            <>
                <Menu
                    borderless
                    size='large'
                    style={
                        showSidebar
                            ? {
                                borderBottom: 'none',
                                marginBottom: '0',
                                borderTop: 'none',
                                height: '51px'
                            }
                            : {borderTop: 'none', height: '52px'}
                    }
                >
                    <Container>
                        <Menu.Item as={Link} to='/'>
                            <img
                                src={logo}
                                alt='logo'
                                style={{marginRight: '0.75em'}}
                            />
                            <div style={{fontSize: '20px'}}>
                                <b>{systemName}</b>
                            </div>
                        </Menu.Item>
                        <Menu.Menu position='right'>
                            <Menu.Item onClick={toggleSidebar}>
                                <Icon name={showSidebar ? 'close' : 'sidebar'}/>
                            </Menu.Item>
                        </Menu.Menu>
                    </Container>
                </Menu>

                {showSidebar ? (
                    <Segment style={{marginTop: 0, borderTop: '0'}}>
                        {/*<Menu secondary vertical style={{ width: '100%', margin: 0 }}>*/}
                        {/*  {renderButtons(true)}*/}
                        {/*  <Menu.Item>*/}
                        {/*    {userState.user ? (*/}
                        {/*      <Button onClick={logout}>注销</Button>*/}
                        {/*    ) : (*/}
                        {/*      <>*/}
                        {/*        <Button*/}
                        {/*          onClick={() => {*/}
                        {/*            setShowSidebar(false);*/}
                        {/*            navigate('/login');*/}
                        {/*          }}*/}
                        {/*        >*/}
                        {/*          登录*/}
                        {/*        </Button>*/}
                        {/*        <Button*/}
                        {/*          onClick={() => {*/}
                        {/*            setShowSidebar(false);*/}
                        {/*            navigate('/register');*/}
                        {/*          }}*/}
                        {/*        >*/}
                        {/*          注册*/}
                        {/*        </Button>*/}
                        {/*      </>*/}
                        {/*    )}*/}
                        {/*  </Menu.Item>*/}
                        {/*</Menu>*/}

                    </Segment>
                ) : (
                    <></>
                )}


            </>
        );
    }

    return (
        <>
            <Layout>
                <div style={{height: '100%'}}>
                    <Nav
                        // mode={'horizontal'}
                        // bodyStyle={{ height: 100 }}
                        selectedKeys={selectedKeys}
                        renderWrapper={({itemElement, isSubNav, isInSubNav, props}) => {
                            const routerMap = {
                                home: "/",
                                channel: "/channel",
                                token: "/token",
                                redemption: "/redemption",
                                topup: "/topup",
                                user: "/user",
                                log: "/log",
                                midjourney: "/midjourney",
                                setting: "/setting",
                                about: "/about",
                            };
                            return (
                                <Link
                                    style={{textDecoration: "none"}}
                                    to={routerMap[props.itemKey]}
                                >
                                    {itemElement}
                                </Link>
                            );
                        }}
                        items={headerButtons}
                        onSelect={key => {
                            console.log(key);
                            setSelectedKeys([key.itemKey]);
                        }}
                        header={{
                            logo: <img src={logo} alt='logo' style={{marginRight: '0.75em'}}/>,
                            text: 'NekoAPI'
                        }}
                        // footer={{
                        //   text: '© 2021 NekoAPI',
                        // }}
                    >

                        <Nav.Footer collapseButton={true}>
                        </Nav.Footer>
                    </Nav>
                </div>
            </Layout>
        </>
    );
};

export default HeaderBar;
