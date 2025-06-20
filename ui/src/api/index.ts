// api/index.ts
import * as announcements from './announcements';
import * as subscription from './subscription'; // 示例：订阅相关API
import * as credits from './credits'; // 积分相关API
import * as auth from './auth'; // 认证相关API
// import * as user from './user'; // 示例：用户相关API

export default {
  announcements,
  subscription,
  credits,
  auth,
  // user,
}; 