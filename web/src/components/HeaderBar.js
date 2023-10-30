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
    IconHelpCircle,
    IconCreditCard,
    IconSemiLogo,
    IconHome,
    IconImage
} from '@douyinfe/semi-icons';
import {Nav, Avatar, Dropdown, Layout, Switch} from '@douyinfe/semi-ui';
import {stringToColor} from "../helpers/render";

// HeaderBar Buttons
let headerButtons = [
    {
        text: '关于',
        itemKey: 'about',
        to: '/about',
        icon: <IconHelpCircle/>
    },
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
                        <Menu secondary vertical style={{ width: '100%', margin: 0 }}>
                          {renderButtons(true)}
                          <Menu.Item>
                            {userState.user ? (
                              <Button onClick={logout}>注销</Button>
                            ) : (
                              <>
                                <Button
                                  onClick={() => {
                                    setShowSidebar(false);
                                    navigate('/login');
                                  }}
                                >
                                  登录
                                </Button>
                                <Button
                                  onClick={() => {
                                    setShowSidebar(false);
                                    navigate('/register');
                                  }}
                                >
                                  注册
                                </Button>
                              </>
                            )}
                          </Menu.Item>
                        </Menu>

                    </Segment>
                ) : (
                    <></>
                )}


            </>
        );
    }
    const switchMode = (model) => {
        const body = document.body;
        if (!model) {
            body.removeAttribute('theme-mode');
        } else {
            body.setAttribute('theme-mode', 'dark');
        }
    };
    return (
        <>
            <Layout>
                <div style={{width: '100%'}}>
                    <Nav
                        mode={'horizontal'}
                        // bodyStyle={{ height: 100 }}
                        renderWrapper={({itemElement, isSubNav, isInSubNav, props}) => {
                            const routerMap = {
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
                        selectedKeys={[]}
                        // items={headerButtons}
                        onSelect={key => console.log(key)}
                        footer={
                            <>
                                <Nav.Item itemKey={'about'} icon={<IconHelpCircle />} />
                                <Switch checkedText="🌞" size={'large'} uncheckedText="🌙" onChange={switchMode} />
                                {userState.user ?
                                    <>
                                        <Dropdown
                                            position="bottomRight"
                                            render={
                                                <Dropdown.Menu>
                                                    <Dropdown.Item onClick={logout}>退出</Dropdown.Item>
                                                </Dropdown.Menu>
                                            }
                                        >
                                            <Avatar size="small" color={stringToColor(userState.user.username)} style={{ margin: 4 }}>
                                                {userState.user.username[0]}
                                            </Avatar>
                                            <span>{userState.user.username}</span>
                                        </Dropdown>
                                    </>
                                    :
                                    <>
                                        <Nav.Item itemKey={'login'} text={'登录'} icon={<IconKey />} />
                                        <Nav.Item itemKey={'register'} text={'注册'} icon={<IconUser />} />
                                    </>
                                }
                            </>
                        }
                    >
                    </Nav>
                </div>
            </Layout>
        </>
    );
};

export default HeaderBar;
