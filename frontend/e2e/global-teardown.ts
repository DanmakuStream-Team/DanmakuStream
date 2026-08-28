import { execSync } from 'node:child_process'
import { rmSync } from 'node:fs'
import path from 'node:path'
import { USERS } from './test-data'

export default async function globalTeardown() {
  if (process.env.E2E_MEMBER_C_RUN !== '1') return

  const mysqlCommand = process.env.MYSQL_CMD
    ?? (process.platform === 'linux' ? 'mysql -S /home/haoyue/dms-mysql.sock -uroot -ppassword danmakustream' : '')
  if (mysqlCommand) {
    const uploadedVideoIDs = execSync(
      `${mysqlCommand} -N -B -e "SELECT id FROM videos WHERE title LIKE 'E2E-MC-%';"`,
      { encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] },
    ).trim().split(/\s+/).filter((value) => /^\d+$/.test(value))
    execSync(
      `${mysqlCommand} -e "
      DELETE FROM video_daily_stats WHERE creator_id=(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}');
      DELETE FROM creator_daily_stats WHERE creator_id=(SELECT id FROM users WHERE nickname='${USERS.memberCCreator.nickname}');
      DELETE FROM videos WHERE title LIKE 'E2E-MC-%';
      DELETE FROM users WHERE nickname IN ('${USERS.memberCCreator.nickname}','${USERS.memberCModerator.nickname}','${USERS.memberCPlain.nickname}');"`,
      { stdio: 'pipe' },
    )
    for (const id of uploadedVideoIDs) {
      rmSync(path.resolve('../backend/data/videos', id), { recursive: true, force: true })
    }
  }
  rmSync('/tmp/danmakustream-member-c-fixtures', { recursive: true, force: true })
  rmSync(path.resolve('../backend/data/videos/e2e-member-c'), { recursive: true, force: true })
}
