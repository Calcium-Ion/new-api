import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Popconfirm, Row, Space, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
  verifyJSONPromise
} from '../../../helpers';

export default function SettingsMagnification(props) {
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    ModelPrice: '',
    ModelRatio: '',
    CompletionRatio: '',
    GroupRatio: '',
    UserUsableGroups: ''
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  async function onSubmit() {
    try {
      console.log('Starting validation...');
      await refForm.current.validate().then(() => {
        console.log('Validation passed');
        const updateArray = compareObjects(inputs, inputsRow);
        if (!updateArray.length) return showWarning('你似乎并没有修改什么');
        const requestQueue = updateArray.map((item) => {
          let value = '';
          if (typeof inputs[item.key] === 'boolean') {
            value = String(inputs[item.key]);
          } else {
            value = inputs[item.key];
          }
          return API.put('/api/option/', {
            key: item.key,
            value
          });
        });
        setLoading(true);
        Promise.all(requestQueue)
          .then((res) => {
            if (requestQueue.length === 1) {
              if (res.includes(undefined)) return;
            } else if (requestQueue.length > 1) {
              if (res.includes(undefined))
                return showError('部分保存失败，请重试');
            }
            for (let i = 0; i < res.length; i++) {
              if (!res[i].data.success) {
                return showError(res[i].data.message)
              }
            }
            showSuccess('保存成功');
            props.refresh();
          })
          .catch(error => {
            console.error('Unexpected error in Promise.all:', error);

            showError('保存失败，请重试');
          })
          .finally(() => {
            setLoading(false);
          });
      }).catch((error) => {
        console.error('Validation failed:', error);
        showError('请检查输入');
      });
    } catch (error) {
      showError('请检查输入');
      console.error(error);
    }
  }

  async function resetModelRatio() {
    try {
      let res = await API.post(`/api/option/rest_model_ratio`);
      // return {success, message}
      if (res.data.success) {
        showSuccess(res.data.message);
        props.refresh();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(error);
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);

  return (
    <Spin spinning={loading}>
      <Form
        values={inputs}
        getFormApi={(formAPI) => (refForm.current = formAPI)}
        style={{ marginBottom: 15 }}
      >
        <Form.Section text={'倍率设置'}>
          <Row gutter={16}>
            <Col span={16}>
              <Form.TextArea
                label={'模型固定价格'}
                extraText={'一次调用消耗多少刀，优先级大于模型倍率'}
                placeholder={
                  '为一个 JSON 文本，键为模型名称，值为一次调用消耗多少刀，比如 "gpt-4-gizmo-*": 0.1，一次消耗0.1刀'
                }
                field={'ModelPrice'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => {
                      return verifyJSON(value);
                    },
                    message: '不是合法的 JSON 字符串'
                  }
                ]}
                onChange={(value) =>
                  setInputs({
                    ...inputs,
                    ModelPrice: value
                  })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={16}>
              <Form.TextArea
                label={'模型倍率'}
                extraText={''}
                placeholder={'为一个 JSON 文本，键为模型名称，值为倍率'}
                field={'ModelRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => {
                      return verifyJSON(value);
                    },
                    message: '不是合法的 JSON 字符串'
                  }
                ]}
                onChange={(value) =>
                  setInputs({
                    ...inputs,
                    ModelRatio: value
                  })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={16}>
              <Form.TextArea
                label={'模型补全倍率（仅对自定义模型有效）'}
                extraText={'仅对自定义模型有效'}
                placeholder={'为一个 JSON 文本，键为模型名称，值为倍率'}
                field={'CompletionRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => {
                      return verifyJSON(value);
                    },
                    message: '不是合法的 JSON 字符串'
                  }
                ]}
                onChange={(value) =>
                  setInputs({
                    ...inputs,
                    CompletionRatio: value
                  })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={16}>
              <Form.TextArea
                label={'分组倍率'}
                extraText={''}
                placeholder={'为一个 JSON 文本，键为分组名称，值为倍率'}
                field={'GroupRatio'}
                autosize={{ minRows: 6, maxRows: 12 }}
                trigger='blur'
                stopValidateWithError
                rules={[
                  {
                    validator: (rule, value) => {
                      return verifyJSON(value);
                    },
                    message: '不是合法的 JSON 字符串'
                  }
                ]}
                onChange={(value) =>
                  setInputs({
                    ...inputs,
                    GroupRatio: value
                  })
                }
              />
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={16}>
              <Form.TextArea
                  label={'用户可选分组'}
                  extraText={''}
                  placeholder={'为一个 JSON 文本，键为分组名称，值为倍率'}
                  field={'UserUsableGroups'}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => {
                        return verifyJSON(value);
                      },
                      message: '不是合法的 JSON 字符串'
                    }
                  ]}
                  onChange={(value) =>
                      setInputs({
                        ...inputs,
                        UserUsableGroups: value
                      })
                  }
              />
            </Col>
          </Row>
        </Form.Section>
      </Form>
      <Space>
        <Button onClick={onSubmit}>
          保存倍率设置
        </Button>
        <Popconfirm
          title='确定重置模型倍率吗？'
          content='此修改将不可逆'
          okType={'danger'}
          position={'top'}
          onConfirm={() => {
            resetModelRatio();
          }}
        >
          <Button type={'danger'}>
            重置模型倍率
          </Button>
        </Popconfirm>
      </Space>
    </Spin>
  );
}
