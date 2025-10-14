# SharePoint & Teams Integration Setup Guide

This guide explains how to configure the new SharePoint documentation and Microsoft Teams webhook features for your Prisma Cloud webhook receiver.

## Overview

When a Prisma Cloud alert arrives, the system now:
1. Creates a ClickUp task (existing functionality)
2. Creates a SharePoint documentation page (new)
3. Sends a Microsoft Teams notification (new)

Both SharePoint and Teams integrations are **optional** and can be enabled independently.

---

## Configuration

Add the following environment variables to your `.env` file or system environment:

### SharePoint Integration (via Microsoft Graph API)

```env
# Azure AD App Registration
AZURE_TENANT_ID=your-tenant-id
AZURE_CLIENT_ID=your-client-id
AZURE_CLIENT_SECRET=your-client-secret

# SharePoint Site
SHAREPOINT_SITE_ID=your-site-id
```

### Microsoft Teams Webhook

```env
# Teams Incoming Webhook URL
TEAMS_WEBHOOK_URL=https://outlook.office.com/webhook/...
```

---

## Azure AD App Setup (for SharePoint)

### 1. Create an Azure AD App Registration

1. Go to [Azure Portal](https://portal.azure.com)
2. Navigate to **Azure Active Directory** → **App registrations** → **New registration**
3. Name: `Prisma-Webhook-SharePoint`
4. Supported account types: **Accounts in this organizational directory only**
5. Redirect URI: Leave blank
6. Click **Register**

### 2. Configure API Permissions

1. Go to **API permissions** in your app
2. Click **Add a permission** → **Microsoft Graph** → **Application permissions**
3. Add: `Sites.ReadWrite.All`
4. Click **Grant admin consent** (requires admin privileges)

### 3. Create a Client Secret

1. Go to **Certificates & secrets**
2. Click **New client secret**
3. Description: `Prisma-Webhook-Secret`
4. Expires: Choose appropriate duration
5. Click **Add**
6. **Copy the secret value immediately** (you won't see it again)

### 4. Get Your Tenant ID and Client ID

- **Tenant ID**: Found in **Overview** → **Directory (tenant) ID**
- **Client ID**: Found in **Overview** → **Application (client) ID**

### 5. Get Your SharePoint Site ID

Run this command (replace values):

```bash
curl -X GET "https://graph.microsoft.com/v1.0/sites/{hostname}:/sites/{site-name}" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

Or use this PowerShell script:

```powershell
$siteUrl = "https://yourtenant.sharepoint.com/sites/yoursite"
$response = Invoke-RestMethod -Uri "https://graph.microsoft.com/v1.0/sites/root:$siteUrl" -Headers @{Authorization="Bearer $accessToken"}
$response.id
```

The Site ID format is: `{hostname},{site-id},{web-id}`

---

## Microsoft Teams Webhook Setup

### 1. Create an Incoming Webhook in Teams

1. Open Microsoft Teams
2. Navigate to the channel where you want alerts
3. Click **⋯** (More options) → **Connectors**
4. Search for **Incoming Webhook**
5. Click **Configure**
6. Name: `Prisma Cloud Alerts`
7. Upload an icon (optional)
8. Click **Create**
9. **Copy the webhook URL**
10. Click **Done**

### 2. Add to Configuration

Add the webhook URL to your `.env` file:

```env
TEAMS_WEBHOOK_URL=https://outlook.office.com/webhook/abc123.../IncomingWebhook/def456.../ghi789...
```

---

## Example Configuration (.env file)

```env
# Existing configuration
PORT=8080
CLICKUP_API_TOKEN=pk_12345...
CLICKUP_LIST_ID=901234567
CLICKUP_ASSIGNEES=12345678,87654321
WEBHOOK_API_KEY=your-secure-api-key
ALLOWED_IPS=203.0.113.1,203.0.113.2

# SharePoint Integration (optional)
AZURE_TENANT_ID=xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
AZURE_CLIENT_ID=yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy
AZURE_CLIENT_SECRET=zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz
SHAREPOINT_SITE_ID=contoso.sharepoint.com,12345678-1234-1234-1234-123456789012,87654321-4321-4321-4321-210987654321

# Teams Webhook (optional)
TEAMS_WEBHOOK_URL=https://outlook.office.com/webhook/...
```

---

## Testing

### 1. Check Service Status

After starting the service, check the logs:

```
SharePoint integration enabled
Teams webhook integration enabled
```

If you see warnings, verify your configuration.

### 2. Send a Test Alert

Use the existing Prisma Cloud test webhook or send a manual POST request to `/webhook`.

### 3. Expected Response

```json
{
  "status": "success",
  "received": 1,
  "tasks_created": 1,
  "task_ids": ["abc123"],
  "sharepoint_pages_created": 1,
  "sharepoint_urls": ["https://contoso.sharepoint.com/sites/yoursite/SitePages/..."],
  "teams_notifications_sent": 1
}
```

---

## Troubleshooting

### SharePoint Issues

**Error: "failed to get access token"**
- Verify `AZURE_TENANT_ID`, `AZURE_CLIENT_ID`, and `AZURE_CLIENT_SECRET`
- Check that the client secret hasn't expired

**Error: "page creation failed (status 403)"**
- Ensure `Sites.ReadWrite.All` permission is granted
- Verify admin consent was given
- Check that the app has access to the specific SharePoint site

**Error: "failed to create page"**
- Verify `SHAREPOINT_SITE_ID` format is correct
- Ensure the site exists and the app has permissions

### Teams Issues

**Error: "Teams webhook failed (status 400)"**
- Verify the webhook URL is correct and complete
- Check that the webhook hasn't been deleted in Teams

**No notification received**
- Verify the webhook URL in `.env`
- Check Teams channel settings
- Ensure the webhook connector is still configured

---

## Features

### SharePoint Page Content

Each alert creates a page with:
- **Title**: `[SEVERITY] - Policy Name`
- **Severity badge**: Color-coded by severity
- **Alert details table**: Policy, severity, resource, account, region, timestamp
- **Description**: Full policy description
- **Recommendation**: Remediation steps
- **Links**: Direct link to the ClickUp task

### Teams Notification

Each alert sends a MessageCard with:
- **Color-coded theme**: Based on severity (red for critical/high, orange for medium, yellow for low)
- **Summary**: Severity and policy name
- **Facts**: Policy, resource, account, cloud, region, timestamp
- **Action buttons**:
  - "View ClickUp Task" (links to task)
  - "View Documentation" (links to SharePoint page)

---

## Security Considerations

1. **Store secrets securely**: Never commit `.env` file to version control
2. **Client secret expiry**: Set calendar reminders to rotate secrets before expiry
3. **Least privilege**: The app only requests `Sites.ReadWrite.All` (no other Graph permissions)
4. **Webhook URL protection**: Keep Teams webhook URL confidential
5. **IP allowlist**: Maintain the existing IP allowlist for webhook endpoint

---

## Maintenance

### Rotate Azure AD Client Secret

1. Create a new client secret in Azure Portal
2. Update `AZURE_CLIENT_SECRET` in your environment
3. Restart the service
4. Delete the old secret after verifying the new one works

### Update Teams Webhook

If you need to change the Teams channel:
1. Create a new webhook in the new channel
2. Update `TEAMS_WEBHOOK_URL`
3. Restart the service

---

## Optional: Disable Features

To disable SharePoint or Teams integration:

**Disable SharePoint**: Remove or leave empty the Azure AD environment variables
**Disable Teams**: Remove or leave empty `TEAMS_WEBHOOK_URL`

The service will continue to work with just ClickUp task creation.

---

## Support

For issues specific to:
- **Microsoft Graph API**: https://docs.microsoft.com/graph
- **SharePoint API**: https://docs.microsoft.com/sharepoint/dev
- **Teams Webhooks**: https://docs.microsoft.com/microsoftteams/platform/webhooks-and-connectors
- **Azure AD**: https://docs.microsoft.com/azure/active-directory
