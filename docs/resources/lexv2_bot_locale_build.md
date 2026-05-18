# awsext_lexv2_bot_locale_build

Triggers a build of an Amazon Lex V2 bot locale using the `BuildBotLocale` API.
Polls until the locale status reaches `Built` or the build fails. Changing any
attribute forces replacement. Delete is a no-op — the bot owns the locale lifecycle.

## Create flow

1. Calls `BuildBotLocale` with the specified `bot_id`, `bot_version`, and `locale_id`.
2. Polls `DescribeBotLocale` every 5 seconds (up to 24 iterations / 2 minutes) until
   `BotLocaleStatus` is `Built`.
3. Treats `Failed` or a `Building` → `NotBuilt` transition as a build failure.
4. On success, sets `build_id` to `"<bot_id>/<bot_version>/<locale_id>/<source_hash>"`.

## Example Usage

```hcl
resource "awsext_lexv2_bot_locale_build" "addservice_en_us" {
  bot_id      = awsext_lexv2_bot_import.addservice.bot_id
  bot_version = "DRAFT"
  locale_id   = "en_US"
  source_hash = "abc123"
}
```

## Argument Reference

### Required

- `bot_id` (String) — The identifier of the Amazon Lex V2 bot that owns the locale. Forces replacement.
- `bot_version` (String) — The version of the bot to build. Typically `"DRAFT"`. Forces replacement.
- `locale_id` (String) — The identifier of the locale to build (e.g. `"en_US"`). Forces replacement.
- `source_hash` (String) — A hash of the bot locale source (intents, slots, etc.). Changing this value triggers a rebuild via replacement.

### Computed

- `build_id` (String) — A synthesized identifier for the build, formatted as `"<bot_id>/<bot_version>/<locale_id>/<source_hash>"`.

## Import

Import an existing bot locale build using the synthesized build ID:

```shell
terraform import awsext_lexv2_bot_locale_build.addservice_en_us <bot_id>/DRAFT/en_US/<source_hash>
```

The import ID must follow the format `<bot_id>/<bot_version>/<locale_id>/<source_hash>`.
