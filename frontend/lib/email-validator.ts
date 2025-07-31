// 支持的邮箱域名列表
export const ALLOWED_EMAIL_DOMAINS = [
  'qq.com',
  'outlook.com', 
  'foxmail.com',
  '163.com',
  'cloxl.com',
  '52ai.org'
];

/**
 * 验证邮箱格式
 */
export const isValidEmail = (email: string): boolean => {
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return emailRegex.test(email);
};

/**
 * 提取邮箱域名
 */
export const getEmailDomain = (email: string): string => {
  const parts = email.split('@');
  return parts.length === 2 ? parts[1].toLowerCase() : '';
};

/**
 * 验证邮箱域名是否在允许列表中
 */
export const isAllowedEmailDomain = (email: string): boolean => {
  if (!isValidEmail(email)) {
    return false;
  }
  
  const domain = getEmailDomain(email);
  return ALLOWED_EMAIL_DOMAINS.includes(domain);
};

/**
 * 获取邮箱验证错误信息
 */
export const getEmailValidationError = (email: string): string | null => {
  if (!email) {
    return '请输入邮箱地址';
  }
  
  if (!isValidEmail(email)) {
    return '请输入有效的邮箱格式';
  }
  
  if (!isAllowedEmailDomain(email)) {
    return `不支持的邮箱域名，仅支持: ${ALLOWED_EMAIL_DOMAINS.join(', ')}`;
  }
  
  return null;
};

/**
 * 获取支持的邮箱域名提示文本
 */
export const getSupportedDomainsText = (): string => {
  return `支持的邮箱: ${ALLOWED_EMAIL_DOMAINS.join(', ')}`;
}; 