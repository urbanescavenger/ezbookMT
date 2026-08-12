package services

import (
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
)

// ImportedDataIdMaps represents the name-to-id maps of the newly created accounts, categories and tags,
// which are used to fill the real ids back into the imported transactions
type ImportedDataIdMaps struct {
	AccountIdMap          map[string]int64
	ExpenseCategoryIdMap  map[string]int64
	IncomeCategoryIdMap   map[string]int64
	TransferCategoryIdMap map[string]int64
	TagIdMap              map[string]int64
}

// CreateImportedData creates the new accounts, categories and tags parsed from the import file,
// and returns the name-to-id maps for filling the real ids back into the imported transactions
func CreateImportedData(c core.Context, accountService *AccountService, categoryService *TransactionCategoryService, tagService *TransactionTagService, uid int64, newAccounts []*models.Account, newExpenseCategories []*models.TransactionCategory, newIncomeCategories []*models.TransactionCategory, newTransferCategories []*models.TransactionCategory, newTags []*models.TransactionTag, clientTimezone *time.Location) *ImportedDataIdMaps {
	idMaps := &ImportedDataIdMaps{
		AccountIdMap:          make(map[string]int64),
		ExpenseCategoryIdMap:  make(map[string]int64),
		IncomeCategoryIdMap:   make(map[string]int64),
		TransferCategoryIdMap: make(map[string]int64),
		TagIdMap:              make(map[string]int64),
	}

	for i := 0; i < len(newAccounts); i++ {
		account := newAccounts[i]
		account.Type = models.ACCOUNT_TYPE_SINGLE_ACCOUNT
		account.ParentAccountId = models.LevelOneAccountParentId
		account.DisplayOrder = 0
		account.Icon = 0
		account.Color = ""
		account.Balance = 0
		account.Hidden = false

		err := accountService.CreateAccounts(c, account, 0, nil, nil, clientTimezone)

		if err != nil {
			log.Errorf(c, "[import_helper.CreateImportedData] failed to create account \"%s\" for user \"uid:%d\", because %s", account.Name, uid, err.Error())
			continue
		}

		idMaps.AccountIdMap[account.Name] = account.AccountId
	}

	createImportedCategories(c, categoryService, uid, newExpenseCategories, idMaps.ExpenseCategoryIdMap)
	createImportedCategories(c, categoryService, uid, newIncomeCategories, idMaps.IncomeCategoryIdMap)
	createImportedCategories(c, categoryService, uid, newTransferCategories, idMaps.TransferCategoryIdMap)

	if len(newTags) > 0 {
		err := tagService.CreateTags(c, uid, newTags, true)

		if err != nil {
			log.Errorf(c, "[import_helper.CreateImportedData] failed to create tags for user \"uid:%d\", because %s", uid, err.Error())
		} else {
			for i := 0; i < len(newTags); i++ {
				tag := newTags[i]
				idMaps.TagIdMap[tag.Name] = tag.TagId
			}
		}
	}

	return idMaps
}

func createImportedCategories(c core.Context, categoryService *TransactionCategoryService, uid int64, newCategories []*models.TransactionCategory, categoryIdMap map[string]int64) {
	for i := 0; i < len(newCategories); i++ {
		category := newCategories[i]
		category.ParentCategoryId = models.LevelOneTransactionCategoryParentId
		category.DisplayOrder = 0
		category.Icon = 0
		category.Color = ""
		category.Hidden = false

		err := categoryService.CreateCategory(c, category)

		if err != nil {
			log.Errorf(c, "[import_helper.createImportedCategories] failed to create category \"%s\" for user \"uid:%d\", because %s", category.Name, uid, err.Error())
			continue
		}

		categoryIdMap[category.Name] = category.CategoryId
	}
}

// ApplyImportedDataIdMaps fills the real ids of the newly created accounts, categories and tags
// back into the imported transactions according to the original names
func ApplyImportedDataIdMaps(transactions models.ImportedTransactionSlice, idMaps *ImportedDataIdMaps) {
	if idMaps == nil {
		return
	}

	for i := 0; i < len(transactions); i++ {
		transaction := transactions[i]

		if accountId, exists := idMaps.AccountIdMap[transaction.OriginalSourceAccountName]; exists {
			transaction.AccountId = accountId
		}

		if accountId, exists := idMaps.AccountIdMap[transaction.OriginalDestinationAccountName]; exists {
			transaction.RelatedAccountId = accountId
		}

		switch transaction.Type {
		case models.TRANSACTION_DB_TYPE_EXPENSE:
			if categoryId, exists := idMaps.ExpenseCategoryIdMap[transaction.OriginalCategoryName]; exists {
				transaction.CategoryId = categoryId
			}
		case models.TRANSACTION_DB_TYPE_INCOME:
			if categoryId, exists := idMaps.IncomeCategoryIdMap[transaction.OriginalCategoryName]; exists {
				transaction.CategoryId = categoryId
			}
		case models.TRANSACTION_DB_TYPE_TRANSFER_OUT:
			if categoryId, exists := idMaps.TransferCategoryIdMap[transaction.OriginalCategoryName]; exists {
				transaction.CategoryId = categoryId
			}
		}

		for j := 0; j < len(transaction.OriginalTagNames); j++ {
			tagName := transaction.OriginalTagNames[j]

			if tagId, exists := idMaps.TagIdMap[tagName]; exists && j < len(transaction.TagIds) {
				transaction.TagIds[j] = utils.Int64ToString(tagId)
			}
		}
	}
}
