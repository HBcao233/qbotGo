from nonebot import on
from nonebot.log import logger
from nonebot.adapters import Bot, Event
from nonebot.adapters.onebot.v11 import MessageSegment
import re
import json
import asyncio

import util
from .data_source import get_twitter, parse_msg, parseMedias


_p = r'^(?:\[CQ:.*\] *)?(?:/?tid) ?(?:https?://)?(?:[a-z]*?(?:twitter|x)\.com/[a-zA-Z0-9_]+/status/)?(\d{13,20})(?:[^0-9a-z\n].*)?$'
pattern = re.compile(_p).search


@on('message').handle()
@on('message_sent').handle()
async def _(bot: Bot, event: Event):
  if not (match := pattern(event.raw_message)):
    return
  if event.message_type != 'group':
    return
  if not (tid := match.group(1)):
    return

  with util.Settings() as data:
    allow_groups = data.setdefault('allow_groups', [])
  if event.group_id not in allow_groups:
    return

  res = await get_twitter(tid)
  if isinstance(res, str):
    return await bot.send_msg(group_id=event.group_id, message=res)
  if 'tombstone' in res.keys():
    logger.info('tombstone: %s', json.dumps(res))
    return await bot.send_msg(
      group_id=event.group_id,
      message=res['tombstone']['text']['text'].replace('了解更多', ''),
    )

  msg = parse_msg(res)
  medias = parseMedias(res)
  if len(medias) == 0:
    return await bot.send_msg(group_id=event.group_id, message=msg)

  messages = [msg, f'https://x.com/i/status/{tid}']
  async with util.curl.Client() as client:
    tasks = [client.getImg(url) for url in medias]
    result = await asyncio.gather(*tasks)
  messages.extend(MessageSegment.image('file://' + i) for i in result)
  await bot.send_group_forward_msg(event.group_id, messages)
