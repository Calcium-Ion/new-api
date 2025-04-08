import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

const CLAUDE_HEADER = {
  'claude-3-7-sonnet-20250219-thinking': {
    'anthropic-beta': [
      'output-128k-2025-02-19',
      'token-efficient-tools-2025-02-19',
    ],
  },
};

const CLAUDE_DEFAULT_MAX_TOKENS = {
  default: 8192,
  'claude-3-haiku-20240307': 4096,
  'claude-3-opus-20240229': 4096,
  'claude-3-7-sonnet-20250219-thinking': 8192,
};

export default function SettingClaudeModel(props) {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    'claude.model_headers_settings': '',
    'claude.thinking_adapter_enabled': true,
    'claude.default_max_tokens': '',
    'claude.thinking_adapter_budget_tokens_percentage': 0.8,
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
    const requestQueue = updateArray.map((item) => {
      let value = String(inputs[item.key]);

      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined))
            return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
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
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('Claude设置')}>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.TextArea
                  label={t('Claude请求头覆盖')}
                  field={'claude.model_headers_settings'}
                  placeholder={
                    t('为一个 JSON 文本，例如：') +
                    '\n' +
                    JSON.stringify(CLAUDE_HEADER, null, 2)
                  }
                  extraText={
                    t('示例') + '\n' + JSON.stringify(CLAUDE_HEADER, null, 2)
                  }
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串'),
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'claude.model_headers_settings': value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.TextArea
                  label={t('缺省 MaxTokens')}
                  field={'claude.default_max_tokens'}
                  placeholder={
                    t('为一个 JSON 文本，例如：') +
                    '\n' +
                    JSON.stringify(CLAUDE_DEFAULT_MAX_TOKENS, null, 2)
                  }
                  extraText={
                    t('示例') +
                    '\n' +
                    JSON.stringify(CLAUDE_DEFAULT_MAX_TOKENS, null, 2)
                  }
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串'),
                    },
                  ]}
                  onChange={(value) =>
                    setInputs({ ...inputs, 'claude.default_max_tokens': value })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col span={16}>
                <Form.Switch
                  label={t('启用Claude思考适配（-thinking后缀）')}
                  field={'claude.thinking_adapter_enabled'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'claude.thinking_adapter_enabled': value,
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Col span={16}>
                {/*//展示MaxTokens和BudgetTokens的计算公式, 并展示实际数字*/}
                <Text>
                  {t(
                    'Claude思考适配 BudgetTokens = MaxTokens * BudgetTokens 百分比',
                  )}
                </Text>
              </Col>
            </Row>
            <Row>
              <Col xs={24} sm={12} md={8} lg={8} xl={8}>
                <Form.InputNumber
                  label={t('思考适配 BudgetTokens 百分比')}
                  field={'claude.thinking_adapter_budget_tokens_percentage'}
                  initValue={''}
                  extraText={t('0.1-1之间的小数')}
                  min={0.1}
                  max={1}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      'claude.thinking_adapter_budget_tokens_percentage': value,
                    })
                  }
                />
              </Col>
            </Row>

            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
