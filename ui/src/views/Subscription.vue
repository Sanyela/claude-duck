<template>
  <div>
    <n-page-header>
      <template #title>
        您的订阅
      </template>
      <template #extra>
        <n-button type="primary" @click="showRedeemModal = true">
          <template #icon><n-icon :component="DocumentTextOutline" /></template>
          兑换优惠码
        </n-button>
      </template>
    </n-page-header>

    <n-alert title="无有效订阅" type="warning" style="margin-top: 16px;" v-if="!activeSubscription && !loadingActiveSubscription">
      您目前没有有效订阅，请购买订阅以使用我们的服务。
    </n-alert>

    <n-card title="活跃订阅" style="margin-top: 16px;" :loading="loadingActiveSubscription">
      <div v-if="activeSubscription">
        <n-thing>
          <template #header>
            {{ activeSubscription.plan.name }}
          </template>
          <template #header-extra>
            <n-tag :type="activeSubscription.status === 'active' ? 'success' : 'warning'">
              {{ activeSubscription.status === 'active' ? '活跃' : activeSubscription.status }}
            </n-tag>
          </template>
          <n-descriptions label-placement="left" bordered :column="1">
            <n-descriptions-item label="价格">
              {{ activeSubscription.plan.pricePerMonth }} {{ activeSubscription.plan.currency }} / 月
            </n-descriptions-item>
            <n-descriptions-item label="主要特性">
              <n-ul>
                <n-li v-for="feature in activeSubscription.plan.features" :key="feature">{{ feature }}</n-li>
              </n-ul>
            </n-descriptions-item>
            <n-descriptions-item label="当前周期结束于">
              {{ new Date(activeSubscription.currentPeriodEnd).toLocaleDateString() }}
            </n-descriptions-item>
            <n-descriptions-item label="将在周期结束时取消">
              {{ activeSubscription.cancelAtPeriodEnd ? '是' : '否' }}
            </n-descriptions-item>
          </n-descriptions>
          <n-space justify="end" style="margin-top: 16px;">
            <n-button type="error" ghost @click="confirmCancelSubscription" v-if="!activeSubscription.cancelAtPeriodEnd">
              取消订阅
            </n-button>
             <n-button type="warning" ghost @click="reactivateSubscription" v-if="activeSubscription.cancelAtPeriodEnd">
              重新激活订阅
            </n-button>
          </n-space>
        </n-thing>
      </div>
      <n-empty description="未找到活跃订阅。" v-else-if="!loadingActiveSubscription">
        <template #extra>
          <n-button type="primary" @click="browsePlans">浏览订阅计划</n-button>
        </template>
      </n-empty>
    </n-card>

    <n-card title="订阅历史" style="margin-top: 16px;" :loading="loadingHistory">
      <n-data-table
        v-if="subscriptionHistory.length > 0"
        :columns="historyColumns"
        :data="subscriptionHistory"
        :pagination="false"
        :bordered="false"
      />
      <n-empty description="未找到订阅历史。" v-else-if="!loadingHistory" />
    </n-card>

    <n-modal v-model:show="showRedeemModal" preset="dialog" title="兑换优惠码">
      <n-input v-model:value="couponCode" placeholder="请输入优惠码" style="margin-top: 16px;" />
      <template #action>
        <n-button @click="showRedeemModal = false">取消</n-button>
        <n-button type="primary" @click="handleRedeem" :loading="redeemingCoupon">确认兑换</n-button>
      </template>
    </n-modal>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import {
  NPageHeader, NCard, NButton, NEmpty, NAlert, NModal, NInput, NIcon, NThing, NTag, NDescriptions, NDescriptionsItem, NUl, NLi, NSpace, NDataTable, useMessage, useDialog
} from 'naive-ui'
import { DocumentTextOutline } from '@vicons/ionicons5'
import type { DataTableColumns } from 'naive-ui'
import {
  getActiveSubscription,
  getSubscriptionHistory,
  redeemCoupon,
  type ActiveSubscription,
  type SubscriptionHistoryItem,
} from '../api/subscription'

const message = useMessage()
const dialog = useDialog()

const activeSubscription = ref<ActiveSubscription | null>(null)
const subscriptionHistory = ref<SubscriptionHistoryItem[]>([])
const loadingActiveSubscription = ref(true)
const loadingHistory = ref(true)
const redeemingCoupon = ref(false)

const showRedeemModal = ref(false)
const couponCode = ref('')

const fetchActiveSubscription = async () => {
  loadingActiveSubscription.value = true
  try {
    const response = await getActiveSubscription()
    activeSubscription.value = response.subscription
  } catch (error) {
    message.error('获取活跃订阅失败')
    console.error(error)
  } finally {
    loadingActiveSubscription.value = false
  }
}

const fetchSubscriptionHistory = async () => {
  loadingHistory.value = true
  try {
    const response = await getSubscriptionHistory()
    subscriptionHistory.value = response.history
  } catch (error) {
    message.error('获取订阅历史失败')
    console.error(error)
  } finally {
    loadingHistory.value = false
  }
}

const historyColumns: DataTableColumns<SubscriptionHistoryItem> = [
  { title: '日期', key: 'date', render: (row) => new Date(row.date).toLocaleDateString() },
  { title: '计划名称', key: 'planName' },
  { title: '金额', key: 'amount', render: (row) => `${row.amount} ${row.currency}` },
  { title: '状态', key: 'status', render: (row) => h(NTag, { type: row.status === 'paid' ? 'success' : 'error' }, { default: () => row.status === 'paid' ? '已支付' : '失败' }) },
  { title: '票据', key: 'invoiceUrl', render: (row) => row.invoiceUrl ? h('a', { href: row.invoiceUrl, target: '_blank' }, '查看') : '-' }
]

onMounted(() => {
  fetchActiveSubscription()
  fetchSubscriptionHistory()
})

const browsePlans = () => {
  message.info('功能待实现：浏览订阅计划')
}

const handleRedeem = async () => {
  if (!couponCode.value.trim()) {
    message.warning('请输入优惠码');
    return;
  }
  redeemingCoupon.value = true;
  try {
    const response = await redeemCoupon({ couponCode: couponCode.value });
    if (response.success) {
      message.success(response.message);
      if (response.newSubscription) {
        activeSubscription.value = response.newSubscription;
      }
      showRedeemModal.value = false;
      couponCode.value = '';
       // 重新获取一下，确保状态最新
      await fetchActiveSubscription();
    } else {
      message.error(response.message || '兑换失败，请检查优惠码是否正确');
    }
  } catch (error) {
    message.error('兑换优惠码时发生错误');
    console.error(error);
  } finally {
    redeemingCoupon.value = false;
  }
}

const confirmCancelSubscription = () => {
  dialog.warning({
    title: '确认取消订阅',
    content: '您确定要取消当前的订阅吗？您的订阅将在当前周期结束后失效。',
    positiveText: '确认取消',
    negativeText: '再想想',
    onPositiveClick: async () => {
      // 模拟取消订阅API调用
      message.loading('正在取消订阅...', { duration: 1500 })
      await new Promise(resolve => setTimeout(resolve, 1500))
      if (activeSubscription.value) {
        activeSubscription.value.cancelAtPeriodEnd = true;
        activeSubscription.value.status = 'canceled'; // 仅为前端演示，真实状态应由后端更新
        message.success('订阅已标记为在周期结束时取消')
      }
    },
  });
}

const reactivateSubscription = async () => {
  // 模拟重新激活API调用
  message.loading('正在重新激活订阅...', { duration: 1500 })
  await new Promise(resolve => setTimeout(resolve, 1500))
   if (activeSubscription.value) {
    activeSubscription.value.cancelAtPeriodEnd = false;
    activeSubscription.value.status = 'active'; // 仅为前端演示
    message.success('订阅已重新激活')
  }
}
</script> 