import { test, expect } from '@playwright/test'

test('critical observatory flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByRole('heading', { name: '声呐概览' })).toBeVisible()

  await page.getByRole('link', { name: 'Topics' }).click()
  await expect(page.getByRole('heading', { name: 'Topics' })).toBeVisible()
  await page.getByRole('button', { name: '创建 Topic' }).click()
  await page.locator('input[placeholder="orders"]').fill('e2e-topic')
  await page.getByRole('button', { name: '保存' }).click()
  await expect(page.getByText('e2e-topic').first()).toBeVisible({ timeout: 8000 })

  await page.getByRole('link', { name: '实验室' }).click()
  await page.getByRole('button', { name: '批量发送' }).click()
  await expect(page.getByText('已写入')).toBeVisible({ timeout: 8000 })

  await page.getByRole('link', { name: '消息浏览' }).click()
  await page.getByRole('button', { name: '查询' }).click()
  await expect(page.locator('table, .empty').first()).toBeVisible()

  await page.getByRole('link', { name: '概览' }).click()
  await expect(page.getByText('消息总量')).toBeVisible()
})
