import {Label} from 'semantic-ui-react';
import {Tag} from "@douyinfe/semi-ui";

export function renderText(text, limit) {
    if (text.length > limit) {
        return text.slice(0, limit - 3) + '...';
    }
    return text;
}

export function renderGroup(group) {
    if (group === '') {
        return <Tag size='large'>default</Tag>;
    }
    let groups = group.split(',');
    groups.sort();
    return <>
        {groups.map((group) => {
            if (group === 'vip' || group === 'pro') {
                return <Tag size='large' color='yellow'>{group}</Tag>;
            } else if (group === 'svip' || group === 'premium') {
                return <Tag size='large' color='red'>{group}</Tag>;
            }
            if (group === 'default') {
                return <Tag size='large'>{group}</Tag>;
            } else {
                return <Tag size='large' color={stringToColor(group)}>{group}</Tag>;
            }
        })}
    </>;
}

export function renderNumber(num) {
    if (num >= 1000000000) {
        return (num / 1000000000).toFixed(1) + 'B';
    } else if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 10000) {
        return (num / 1000).toFixed(1) + 'k';
    } else {
        return num;
    }
}

export function renderQuotaNumberWithDigit(num, digits = 2) {
    let displayInCurrency = localStorage.getItem('display_in_currency');
    num = num.toFixed(digits);
    if (displayInCurrency) {
        return '$' + num;
    }
    return num;
}

export function renderNumberWithPoint(num) {
    num = num.toFixed(2);
    if (num >= 100000) {
        // Convert number to string to manipulate it
        let numStr = num.toString();
        // Find the position of the decimal point
        let decimalPointIndex = numStr.indexOf('.');

        let wholePart = numStr;
        let decimalPart = '';

        // If there is a decimal point, split the number into whole and decimal parts
        if (decimalPointIndex !== -1) {
            wholePart = numStr.slice(0, decimalPointIndex);
            decimalPart = numStr.slice(decimalPointIndex);
        }

        // Take the first two and last two digits of the whole number part
        let shortenedWholePart = wholePart.slice(0, 2) + '..' + wholePart.slice(-2);

        // Return the formatted number
        return shortenedWholePart + decimalPart;
    }

    // If the number is less than 100,000, return it unmodified
    return num;
}

export function getQuotaPerUnit() {
    let quotaPerUnit = localStorage.getItem('quota_per_unit');
    quotaPerUnit = parseFloat(quotaPerUnit);
    return quotaPerUnit;
}

export function getQuotaWithUnit(quota, digits = 6) {
    let quotaPerUnit = localStorage.getItem('quota_per_unit');
    quotaPerUnit = parseFloat(quotaPerUnit);
    return (quota / quotaPerUnit).toFixed(digits);
}

export function renderQuota(quota, digits = 2) {
    let quotaPerUnit = localStorage.getItem('quota_per_unit');
    let displayInCurrency = localStorage.getItem('display_in_currency');
    quotaPerUnit = parseFloat(quotaPerUnit);
    displayInCurrency = displayInCurrency === 'true';
    if (displayInCurrency) {
        return '$' + (quota / quotaPerUnit).toFixed(digits);
    }
    return renderNumber(quota);
}

export function renderQuotaWithPrompt(quota, digits) {
    let displayInCurrency = localStorage.getItem('display_in_currency');
    displayInCurrency = displayInCurrency === 'true';
    if (displayInCurrency) {
        return `（等价金额：${renderQuota(quota, digits)}）`;
    }
    return '';
}

const colors = ['amber', 'blue', 'cyan', 'green', 'grey', 'indigo',
    'light-blue', 'lime', 'orange', 'pink',
    'purple', 'red', 'teal', 'violet', 'yellow'
]

export function stringToColor(str) {
    let sum = 0;
    // 对字符串中的每个字符进行操作
    for (let i = 0; i < str.length; i++) {
        // 将字符的ASCII值加到sum中
        sum += str.charCodeAt(i);
    }
    // 使用模运算得到个位数
    let i = sum % colors.length;
    return colors[i];
}