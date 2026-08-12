<template>
    <v-dialog width="640" :persistent="true" v-model="showState">
        <one-column-dialog-layout :disabled="merging"
                                  :title="tt('Merge into another account')"
                                  :cancel-button-title="tt('Cancel')"
                                  @cancel="cancel">
            <template #content>
                <div class="text-body-large">{{ tt('format.misc.mergeAccountTip', { fromAccount: fromAccount?.name, toAccount: displayToAccountName }) }}</div>
                <div class="w-100 d-flex justify-center mt-6">
                    <v-row>
                        <v-col cols="12" md="12">
                            <two-column-select primary-key-field="id" primary-value-field="category"
                                               primary-title-field="name" primary-footer-field="displayBalance"
                                               primary-icon-field="icon" primary-icon-type="account"
                                               primary-sub-items-field="accounts"
                                               :primary-title-i18n="true"
                                               secondary-key-field="id" secondary-value-field="id"
                                               secondary-title-field="name" secondary-footer-field="displayBalance"
                                               secondary-icon-field="icon" secondary-icon-type="account" secondary-color-field="color"
                                               :disabled="loading || merging || !allVisibleAccounts.length"
                                               :enable-filter="true" :filter-placeholder="tt('Find account')" :filter-no-items-text="tt('No available account')"
                                               :label="tt('Target Account')"
                                               :placeholder="tt('Target Account')"
                                               :items="allVisibleCategorizedAccounts"
                                               :no-item-text="Account.findAccountNameById(allAccounts, toAccountId, tt('Unspecified'))"
                                               v-model="toAccountId">
                            </two-column-select>
                        </v-col>

                        <v-col cols="12" md="12">
                            <v-text-field type="text"
                                          persistent-placeholder
                                          :disabled="merging"
                                          :label="tt('Confirm Target Account Name')"
                                          :placeholder="tt('Please re-enter the target account name to confirm')"
                                          v-model="toAccountName"
                            />
                        </v-col>
                    </v-row>
                </div>
            </template>

            <template #footer>
                <v-btn color="secondary" variant="tonal" :disabled="merging" @click="cancel">{{ tt('Cancel') }}</v-btn>
                <v-spacer/>
                <v-btn :disabled="!fromAccount || !toAccountId || fromAccount?.id === toAccountId || !toAccountName || !isToAccountNameValid || merging" @click="confirm">
                    {{ tt('Confirm') }}
                    <v-progress-circular indeterminate size="22" class="ms-2" v-if="merging"></v-progress-circular>
                </v-btn>
            </template>
        </one-column-dialog-layout>
    </v-dialog>

    <snack-bar ref="snackbar" />
</template>

<script setup lang="ts">
import SnackBar from '@/components/desktop/SnackBar.vue';

import { ref, useTemplateRef } from 'vue';

import { useI18n } from '@/locales/helpers.ts';
import { useMergeAccountPageBase } from '@/views/base/accounts/MergeAccountPageBase.ts'

import { useAccountsStore } from '@/stores/account.ts';
import { useTransactionsStore } from '@/stores/transaction.ts';

import { Account } from '@/models/account.ts';

type SnackBarType = InstanceType<typeof SnackBar>;

const { tt } = useI18n();

const {
    merging,
    fromAccount,
    toAccountId,
    toAccountName,
    allAccounts,
    allVisibleAccounts,
    allVisibleCategorizedAccounts,
    displayToAccountName,
    isToAccountNameValid
} = useMergeAccountPageBase();

const accountsStore = useAccountsStore();
const transactionsStore = useTransactionsStore();

let resolveFunc: (() => void) | null = null;
let rejectFunc: ((reason?: unknown) => void) | null = null;

const snackbar = useTemplateRef<SnackBarType>('snackbar');

const showState = ref<boolean>(false);
const loading = ref<boolean>(false);

function init(): void {
    accountsStore.loadAllAccounts({
        force: false
    }).then(() => {
        loading.value = false;
    }).catch(error => {
        loading.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function open(account: Account): Promise<void> {
    showState.value = true;
    merging.value = false;
    fromAccount.value = account;
    toAccountId.value = '';
    toAccountName.value = '';

    return new Promise((resolve, reject) => {
        resolveFunc = resolve;
        rejectFunc = reject;
    });
}

function confirm(): void {
    if (!fromAccount.value || !toAccountId.value || fromAccount.value?.id === toAccountId.value || !toAccountName.value || !isToAccountNameValid.value) {
        return;
    }

    merging.value = true;

    transactionsStore.mergeAccounts({
        targetAccountId: toAccountId.value,
        mergedAccountIds: [fromAccount.value.id]
    }).then(() => {
        merging.value = false;

        resolveFunc?.();
        showState.value = false;
    }).catch(error => {
        merging.value = false;

        if (!error.processed) {
            snackbar.value?.showError(error);
        }
    });
}

function cancel(): void {
    rejectFunc?.();
    showState.value = false;
}

defineExpose({
    open
});

init();
</script>
