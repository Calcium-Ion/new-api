import React, {useEffect, useState} from 'react';
import {Button, Card, Confirm, Form, Grid, Header, Segment, Statistic} from 'semantic-ui-react';
import {API, isMobile, showError, showInfo, showSuccess} from '../../helpers';
import {renderNumber, renderQuota} from '../../helpers/render';
import {Col, Layout, Row, Typography} from "@douyinfe/semi-ui";

const TopUp = () => {
    const [redemptionCode, setRedemptionCode] = useState('');
    const [topUpCode, setTopUpCode] = useState('');
    const [topUpCount, setTopUpCount] = useState(10);
    const [amount, setAmount] = useState(0);
    const [topUpLink, setTopUpLink] = useState('');
    const [userQuota, setUserQuota] = useState(0);
    const [isSubmitting, setIsSubmitting] = useState(false);
    const [open, setOpen] = useState(false);
    const [payWay, setPayWay] = useState('');

    const topUp = async () => {
        if (redemptionCode === '') {
            showInfo('请输入充值码！')
            return;
        }
        setIsSubmitting(true);
        try {
            const res = await API.post('/api/user/topup', {
                key: redemptionCode
            });
            const {success, message, data} = res.data;
            if (success) {
                showSuccess('充值成功！');
                setUserQuota((quota) => {
                    return quota + data;
                });
                setRedemptionCode('');
            } else {
                showError(message);
            }
        } catch (err) {
            showError('请求失败');
        } finally {
            setIsSubmitting(false);
        }
    };

    const openTopUpLink = () => {
        if (!topUpLink) {
            showError('超级管理员未设置充值链接！');
            return;
        }
        window.open(topUpLink, '_blank');
    };

    const preTopUp = async (payment) => {
        if (amount === 0) {
            await getAmount();
        }
        setPayWay(payment)
        setOpen(true);
    }

    const onlineTopUp = async () => {
        if (amount === 0) {
            await getAmount();
        }
        setOpen(false);
        try {
            const res = await API.post('/api/user/pay', {
                amount: parseInt(topUpCount),
                top_up_code: topUpCode,
                PaymentMethod: payWay
            });
            if (res !== undefined) {
                const {message, data} = res.data;
                // showInfo(message);
                if (message === 'success') {
                    let params = data
                    let url = res.data.url
                    let form = document.createElement('form')
                    form.action = url
                    form.method = 'POST'
                    form.target = '_blank'
                    for (let key in params) {
                        let input = document.createElement('input')
                        input.type = 'hidden'
                        input.name = key
                        input.value = params[key]
                        form.appendChild(input)
                    }
                    document.body.appendChild(form)
                    form.submit()
                    document.body.removeChild(form)
                } else {
                    showError(data);
                    // setTopUpCount(parseInt(res.data.count));
                    // setAmount(parseInt(data));
                }
            } else {
                showError(res);
            }
        } catch (err) {
            console.log(err);
        } finally {
        }
    }

    const getUserQuota = async () => {
        let res = await API.get(`/api/user/self`);
        const {success, message, data} = res.data;
        if (success) {
            setUserQuota(data.quota);
        } else {
            showError(message);
        }
    }

    useEffect(() => {
        let status = localStorage.getItem('status');
        if (status) {
            status = JSON.parse(status);
            if (status.top_up_link) {
                setTopUpLink(status.top_up_link);
            }
        }
        getUserQuota().then();
    }, []);

    const renderAmount = () => {
        console.log(amount);
        return amount + '元';
    }

    const getAmount = async (value) => {
        if (value === undefined) {
            value = topUpCount;
        }
        try {
            const res = await API.post('/api/user/amount', {
                amount: parseFloat(value),
                top_up_code: topUpCode
            });
            if (res !== undefined) {
                const {message, data} = res.data;
                // showInfo(message);
                if (message === 'success') {
                    setAmount(parseInt(data));
                } else {
                    showError(data);
                    // setTopUpCount(parseInt(res.data.count));
                    // setAmount(parseInt(data));
                }
            } else {
                showError(res);
            }
        } catch (err) {
            console.log(err);
        } finally {
        }
    }

    const handleCancel = () => {
        setOpen(false);
    }

    return (
        <div>
            <Layout>
                <Layout.Header>
                    <h3>我的钱包</h3>
                </Layout.Header>
                <Layout.Content>
                    <div style={{marginTop: 20, paddingLeft: isMobile()?0:200, paddingRight: isMobile()?0:200}}>
                        <Card
                            style={{width: '100%', padding: '20px'}}
                        >
                            <Confirm
                                open={open}
                                content={'充值数量：' + topUpCount + '，充值金额：' + renderAmount() + '，是否确认充值？'}
                                cancelButton='取消充值'
                                confirmButton="确定"
                                onCancel={handleCancel}
                                onConfirm={onlineTopUp}
                            />
                            <h3>兑换</h3>
                            <Grid columns={2} stackable>
                                <Grid.Column>
                                    <Form>
                                        <Form.Input
                                            placeholder='兑换码'
                                            name='redemptionCode'
                                            value={redemptionCode}
                                            onChange={(e) => {
                                                setRedemptionCode(e.target.value);
                                            }}
                                        />
                                        <Button color='green' onClick={openTopUpLink}>
                                            获取兑换码
                                        </Button>
                                        <Button color='yellow' onClick={topUp} disabled={isSubmitting}>
                                            {isSubmitting ? '兑换中...' : '兑换'}
                                        </Button>
                                    </Form>
                                </Grid.Column>
                                <Grid.Column>
                                    <Statistic.Group widths='one'>
                                        <Statistic>
                                            <Statistic.Value>{renderQuota(userQuota)}</Statistic.Value>
                                            <Statistic.Label>剩余额度</Statistic.Label>
                                        </Statistic>
                                    </Statistic.Group>
                                </Grid.Column>
                            </Grid>
                        </Card>
                        <Card
                            style={{width: '100%', padding: '20px'}}
                        >
                            <Header as='h3'>在线充值，最低1</Header>
                            <Grid columns={2} stackable>
                                <Grid.Column>
                                    <Form>
                                        <Form.Input
                                            placeholder='充值金额，最低1'
                                            name='redemptionCount'
                                            type={'number'}
                                            value={topUpCount}
                                            autoComplete={'off'}
                                            onChange={async (e) => {
                                                setTopUpCount(e.target.value);
                                                await getAmount(e.target.value);
                                            }}
                                        />
                                        <Button color='blue' onClick={
                                            async () => {
                                                preTopUp('zfb')
                                            }
                                        }>
                                            支付宝
                                        </Button>
                                        <Button color='green' onClick={
                                            async () => {
                                                preTopUp('wx')
                                            }
                                        }>
                                            微信
                                        </Button>
                                    </Form>
                                </Grid.Column>
                                <Grid.Column>
                                    <Statistic.Group widths='one'>
                                        <Statistic>
                                            <Statistic.Value>{renderAmount()}</Statistic.Value>
                                            <Statistic.Label>支付金额</Statistic.Label>
                                        </Statistic>
                                    </Statistic.Group>
                                </Grid.Column>
                            </Grid>
                        </Card>
                    </div>

                </Layout.Content>
            </Layout>
        </div>

    );
};

export default TopUp;