import React, {useEffect, useRef, useState} from 'react';
// import {Button, Form, Header, Message, Segment} from 'semantic-ui-react';
import {useParams, useNavigate} from 'react-router-dom';
import {API, isMobile, showError, showSuccess, timestamp2string} from '../../helpers';
import {renderQuota, renderQuotaWithPrompt} from '../../helpers/render';
import {Layout, SideSheet, Button, Space, Spin, Banner, Input, DatePicker, AutoComplete, Typography} from "@douyinfe/semi-ui";
import Title from "@douyinfe/semi-ui/lib/es/typography/title";
import {Divider} from "semantic-ui-react";

const EditToken = (props) => {
    const isEdit = props.editingToken.id !== undefined;
    const [loading, setLoading] = useState(isEdit);
    const originInputs = {
        name: '',
        remain_quota: isEdit ? 0 : 500000,
        expired_time: -1,
        unlimited_quota: false
    };
    const [inputs, setInputs] = useState(originInputs);
    const {name, remain_quota, expired_time, unlimited_quota} = inputs;
    // const [visible, setVisible] = useState(false);
    const navigate = useNavigate();
    const handleInputChange = (name, value) => {
        setInputs((inputs) => ({...inputs, [name]: value}));
    };
    const handleCancel = () => {
        props.handleClose();
    }
    const setExpiredTime = (month, day, hour, minute) => {
        let now = new Date();
        let timestamp = now.getTime() / 1000;
        let seconds = month * 30 * 24 * 60 * 60;
        seconds += day * 24 * 60 * 60;
        seconds += hour * 60 * 60;
        seconds += minute * 60;
        if (seconds !== 0) {
            timestamp += seconds;
            setInputs({...inputs, expired_time: timestamp2string(timestamp)});
        } else {
            setInputs({...inputs, expired_time: -1});
        }
    };

    const setUnlimitedQuota = () => {
        setInputs({...inputs, unlimited_quota: !unlimited_quota});
    };

    const loadToken = async () => {
        setLoading(true);
        let res = await API.get(`/api/token/${props.editingToken.id}`);
        const {success, message, data} = res.data;
        if (success) {
            if (data.expired_time !== -1) {
                data.expired_time = timestamp2string(data.expired_time);
            }
            setInputs(data);
        } else {
            showError(message);
        }
        setLoading(false);
    };
    useEffect(() => {
        if (isEdit) {
            loadToken().then(
                () => {
                    console.log(inputs);
                }
            );
        } else {
            setInputs(originInputs);
        }
    }, [props.editingToken.id]);

    const submit = async () => {
        setLoading(true);
        if (!isEdit && inputs.name === '') return;
        let localInputs = inputs;
        localInputs.remain_quota = parseInt(localInputs.remain_quota);
        if (localInputs.expired_time !== -1) {
            let time = Date.parse(localInputs.expired_time);
            if (isNaN(time)) {
                showError('过期时间格式错误！');
                return;
            }
            localInputs.expired_time = Math.ceil(time / 1000);
        }
        let res;
        if (isEdit) {
            res = await API.put(`/api/token/`, {...localInputs, id: parseInt(props.editingToken.id)});
        } else {
            res = await API.post(`/api/token/`, localInputs);
        }
        const {success, message} = res.data;
        if (success) {
            if (isEdit) {
                showSuccess('令牌更新成功！');
                props.refresh();
                props.handleClose();
            } else {
                showSuccess('令牌创建成功，请在列表页面点击复制获取令牌！');
                setInputs(originInputs);
                props.refresh();
                props.handleClose();
            }
        } else {
            showError(message);
        }
        setLoading(false);
    };


    return (
        <>
            <SideSheet
                title={<Title level={3}>{isEdit ? '更新令牌信息' : '创建新的令牌'}</Title>}
                headerStyle={{borderBottom: '1px solid var(--semi-color-border)'}}
                bodyStyle={{borderBottom: '1px solid var(--semi-color-border)'}}
                visible={props.visiable}
                footer={
                    <div style={{display: 'flex', justifyContent: 'flex-end'}}>
                        <Space>
                            <Button theme='solid' size={'large'} onClick={submit}>提交</Button>
                            <Button theme='solid' size={'large'} type={'tertiary'} onClick={handleCancel}>取消</Button>
                        </Space>
                    </div>
                }
                closeIcon={null}
                onCancel={() => handleCancel()}
                width={isMobile() ? '100%' : 600}
            >
                <Spin spinning={loading}>
                    <Input
                        style={{ marginTop: 20 }}
                        label='名称'
                        name='name'
                        placeholder={'请输入名称'}
                        onChange={(value) => handleInputChange('name', value)}
                        value={name}
                        autoComplete='new-password'
                        required={!isEdit}
                    />
                    <Divider/>
                    <DatePicker
                        label='过期时间'
                        name='expired_time'
                        placeholder={'请选择过期时间'}
                        onChange={(value) => handleInputChange('expired_time', value)}
                        value={expired_time}
                        autoComplete='new-password'
                        type='dateTime'
                    />
                    <div style={{ marginTop: 20 }}>
                        <Space>
                            <Button type={'tertiary'} onClick={() => {
                                setExpiredTime(0, 0, 0, 0);
                            }}>永不过期</Button>
                            <Button type={'tertiary'} onClick={() => {
                                setExpiredTime(0, 0, 1, 0);
                            }}>一小时</Button>
                            <Button type={'tertiary'} onClick={() => {
                                setExpiredTime(1, 0, 0, 0);
                            }}>一个月</Button>
                            <Button type={'tertiary'} onClick={() => {
                                setExpiredTime(0, 1, 0, 0);
                            }}>一天</Button>
                        </Space>
                    </div>

                    <Divider/>
                    <Banner type={'warning'} description={'注意，令牌的额度仅用于限制令牌本身的最大额度使用量，实际的使用受到账户的剩余额度限制。'}></Banner>
                    <div style={{ marginTop: 20 }}>
                        <Typography.Text>{`额度${renderQuotaWithPrompt(remain_quota)}`}</Typography.Text>
                    </div>
                    <AutoComplete
                        style={{ marginTop: 8 }}
                        name='remain_quota'
                        placeholder={'请输入额度'}
                        onChange={(value) => handleInputChange('remain_quota', value)}
                        value={remain_quota}
                        autoComplete='new-password'
                        type='number'
                        position={'top'}
                        data={[
                            {value: 500000, label: '1$'},
                            {value: 5000000, label: '10$'},
                            {value: 25000000, label: '50$'},
                            {value: 50000000, label: '100$'},
                            {value: 250000000, label: '500$'},
                            {value: 500000000, label: '1000$'},
                        ]}
                        disabled={unlimited_quota}
                    />
                    <div>
                        <Button style={{ marginTop: 8 }} type={'warning'} onClick={() => {
                            setUnlimitedQuota();
                        }}>{unlimited_quota ? '取消无限额度' : '设为无限额度'}</Button>
                    </div>
                </Spin>
            </SideSheet>
        </>
    );
};

export default EditToken;
