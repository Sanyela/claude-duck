<template>
  <div>
    <n-page-header>
      <template #title>
        积分
      </template>
    </n-page-header>

    <n-alert title="无有效订阅" type="warning" style="margin-top: 16px;" v-if="!hasActiveSubscription">
      您目前没有有效订阅，部分积分功能（如自动补充、高级模型使用）可能受限或不可用。
    </n-alert>

    <n-grid cols="1 m:2" :x-gap="16" :y-gap="16" style="margin-top: 16px;">
      <n-gi>
        <n-card title="当前积分余额" :loading="loadingBalance">
          <div v-if="creditBalance">
            <n-statistic label="可用积分" :value="creditBalance.available">
              <template #suffix>
                / {{ creditBalance.total }}
              </template>
            </n-statistic>
            <n-text depth="3" style="font-size: 12px; margin-top: 4px;">
              补充速率: {{ creditBalance.rechargeRatePerHour }} 积分/小时
            </n-text>
            <n-button 
              block 
              type="primary" 
              style="margin-top: 16px;" 
              @click="handleRequestCreditReset" 
              :disabled="!creditBalance.canRequestReset || resettingCredits"
              :loading="resettingCredits"
            >
              {{ creditBalance.canRequestReset ? '申请重置积分' : (creditBalance.nextResetTime ? `下次可重置: ${new Date(creditBalance.nextResetTime).toLocaleTimeString()}` : '今日已申请') }}
            </n-button>
          </div>
          <n-empty v-else description="未能加载积分余额" />
        </n-card>
      </n-gi>
      <n-gi>
        <n-card title="模型使用成本" :loading="loadingCosts">
          <n-space vertical v-if="modelCosts.length > 0">
            <n-thing v-for="cost in modelCosts" :key="cost.id">
              <template #header>
                {{ cost.modelName }}
              </template>
              <template #description>
                <n-tag 
                  :type="cost.status === 'available' ? 'success' : (cost.status === 'limited' ? 'warning' : 'error')" 
                  size="small"
                >
                  {{ cost.status === 'available' ? '可用' : (cost.status === 'limited' ? '受限' : '不可用') }}
                </n-tag>
                <span v-if="cost.costFactor" style="margin-left: 8px;"> - 成本因子: {{ cost.costFactor }}x</span>
              </template>
              <p v-if="cost.description" style="font-size: 12px; color: #888;">{{ cost.description }}</p>
            </n-thing>
          </n-space>
          <n-empty v-else description="未能加载模型成本信息" />
        </n-card>
      </n-gi>
    </n-grid>

    <n-card title="积分使用历史" style="margin-top: 16px;" :loading="loadingUsageHistory">
      <n-data-table
        v-if="creditUsageHistory.length > 0"
        :columns="usageHistoryColumns"
        :data="creditUsageHistory"
        :pagination="pagination"
        :remote="true"
        @update:page="handlePageChange"
        :bordered="false"
      />
      <n-empty v-else-if="!loadingUsageHistory" description="暂无积分使用历史" />
    </n-card>

  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, reactive, h } from 'vue';
import {
  NPageHeader, NCard, NGrid, NGi, NStatistic, NButton, NAlert, NEmpty, NText, NTag, NSpace, NThing, NDataTable, useMessage
} from 'naive-ui';
import type { DataTableColumns, PaginationProps } from 'naive-ui';
import {
  getCreditBalance,
  getModelCosts,
  getCreditUsageHistory,
  requestCreditReset,
  type CreditBalance,
  type ModelCost,
  type CreditUsageRecord,
} from '../api/credits';
import { getActiveSubscription, type ActiveSubscription } from '../api/subscription'; // 用于判断订阅状态

const message = useMessage();

const creditBalance = ref<CreditBalance | null>(null);
const modelCosts = ref<ModelCost[]>([]);
const creditUsageHistory = ref<CreditUsageRecord[]>([]);
const activeSubscription = ref<ActiveSubscription | null>(null); // 用于模拟订阅对积分的影响

const loadingBalance = ref(true);
const loadingCosts = ref(true);
const loadingUsageHistory = ref(true);
const resettingCredits = ref(false);

const pagination = reactive<PaginationProps>({
  page: 1,
  pageSize: 10,
  itemCount: 0,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  onChange: (page: number) => {
    pagination.page = page;
    fetchUsageHistory();
  },
  onUpdatePageSize: (pageSize: number) => {
    pagination.pageSize = pageSize;
    pagination.page = 1;
    fetchUsageHistory();
  },
});

const hasActiveSubscription = computed(() => !!activeSubscription.value && activeSubscription.value.status === 'active');

const fetchCreditBalanceData = async () => {
  loadingBalance.value = true;
  try {
    const response = await getCreditBalance();
    creditBalance.value = response.balance;
    // 模拟订阅影响
    if (hasActiveSubscription.value && creditBalance.value) {
        creditBalance.value.rechargeRatePerHour = 100; 
    }
  } catch (error) {
    message.error('获取积分余额失败');
    console.error(error);
  } finally {
    loadingBalance.value = false;
  }
};

const fetchModelCostsData = async () => {
  loadingCosts.value = true;
  try {
    const response = await getModelCosts();
    modelCosts.value = response.costs;
    // 模拟订阅影响
    if (hasActiveSubscription.value) {
        const claudeCode = modelCosts.value.find(c => c.id === 'claude-code');
        if(claudeCode) claudeCode.status = 'available';
    }
  } catch (error) {
    message.error('获取模型成本失败');
    console.error(error);
  } finally {
    loadingCosts.value = false;
  }
};

const fetchUsageHistory = async () => {
  loadingUsageHistory.value = true;
  try {
    const response = await getCreditUsageHistory(pagination.page, pagination.pageSize);
    creditUsageHistory.value = response.history;
    pagination.itemCount = response.totalPages * (pagination.pageSize || 10); // Mock中totalPages代表总页数
  } catch (error) {
    message.error('获取积分使用历史失败');
    console.error(error);
  } finally {
    loadingUsageHistory.value = false;
  }
};

const handleRequestCreditReset = async () => {
  resettingCredits.value = true;
  try {
    const response = await requestCreditReset();
    if (response.success) {
      message.success(response.message);
      await fetchCreditBalanceData(); // 刷新余额
    } else {
      message.warning(response.message + (response.nextAvailableTime ? `下次可申请时间: ${new Date(response.nextAvailableTime).toLocaleTimeString()}`: '' ));
    }
  } catch (error) {
    message.error('申请重置积分失败');
    console.error(error);
  } finally {
    resettingCredits.value = false;
  }
};

const usageHistoryColumns: DataTableColumns<CreditUsageRecord> = [
  { title: '时间', key: 'timestamp', render: (row) => new Date(row.timestamp).toLocaleString(), width: 180 },
  { title: '描述', key: 'description' },
  { title: '相关模型', key: 'relatedModel', width: 150 },
  { title: '积分消耗', key: 'amount', width: 100, render(row) {
      return h(NTag, { type: row.amount < 0 ? 'error' : 'success' }, { default: () => String(row.amount)})
    }
  },
];

const handlePageChange = (page: number) => {
  pagination.page = page;
  fetchUsageHistory();
};

const fetchInitialData = async () => {
    try {
        const subResponse = await getActiveSubscription();
        activeSubscription.value = subResponse.subscription;
    } catch (e) {
        console.error("Failed to fetch initial subscription status for credits page", e);
    }
    await Promise.all([
        fetchCreditBalanceData(),
        fetchModelCostsData(),
        fetchUsageHistory(),
    ]);
};

onMounted(() => {
  fetchInitialData();
});

</script> 