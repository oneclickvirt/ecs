import { spawn } from 'node:child_process'
import { readdir, readFile } from 'node:fs/promises'
import { menuDriver } from './dsh-options.mjs'
import { GoECSError, capture } from './dsh-platform.mjs'

async function groupActive(pid) {
  if (process.platform === 'linux') {
    for (const name of await readdir('/proc')) {
      if (!/^\d+$/.test(name)) continue
      try {
        const record = await readFile('/proc/' + name + '/stat', 'utf8'), fields = record.slice(record.lastIndexOf(')') + 2).split(' ')
        if (Number(fields[2]) === pid && fields[0] !== 'Z') return true
      } catch (error) { if (!['ENOENT','ESRCH'].includes(error.code)) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Cannot inspect local process group') }
    }
    return false
  }
  const processes = await capture(['ps', '-axo', 'pgid=,stat='], { timeout: 2000 })
  if (processes.code !== 0) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Could not confirm local process termination')
  return processes.stdout.split('\n').some(line => { const [group, state] = line.trim().split(/\s+/); return Number(group) === pid && state && !state.startsWith('Z') })
}

export async function startLocalProcess(binary, args, { cwd, env, choice, onData, deadlineMs = 3600000, onDeadline = () => {} }) {
  let pid, write, terminate, done
  if (choice !== 'custom') {
    let pty
    try { pty = await import('@lydell/node-pty') } catch { throw new GoECSError('PTY_UNAVAILABLE', 'A terminal provider is required to run original GoECS menu presets on this system') }
    const terminal = (pty.spawn || pty.default?.spawn)(binary, args, { name: 'xterm-256color', cols: 140, rows: 48, cwd, env })
    pid = terminal.pid; write = data => terminal.write(data)
    const drive = menuDriver(choice, write)
    const dataListener = terminal.onData(data => {
      // Bubble Tea's terminal queries need a terminal response even in CI.
      if (data.includes('\x1b[6n')) write('\x1b[1;1R')
      if (data.includes('\x1b]11;?')) write('\x1b]11;rgb:1111/1818/2727\x1b\\')
      onData(data); drive(data)
    })
    done = new Promise((resolve, reject) => {
      const exitListener = terminal.onExit(result => {
        dataListener.dispose(); exitListener.dispose()
        // ConPTY keeps its output worker alive until the terminal is released.
        try { if (process.platform === 'win32') terminal.kill(); resolve(result) } catch (error) { reject(error) }
      })
    })
    terminate = () => terminal.kill()
  } else {
    const child = spawn(binary, args, { cwd, env, detached: process.platform !== 'win32', windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'] })
    pid = child.pid; write = data => child.stdin.write(data)
    child.stdout.on('data', data => onData(data.toString())); child.stderr.on('data', data => onData(data.toString()))
    done = new Promise((resolve, reject) => { child.once('error', reject); child.once('close', (exitCode, signal) => resolve({ exitCode, signal })) })
    terminate = () => child.kill('SIGTERM')
  }
  let finished = false, cancellation, rejectStop
  const stopFailure = new Promise((_resolve, reject) => { rejectStop = reject })
  const stop = async () => {
    if (cancellation) return cancellation
    if (finished) return done
    cancellation = (async () => {
      if (process.platform === 'win32') {
        const result = await capture(['taskkill.exe', '/PID', String(pid), '/T', '/F'], { timeout: 10000 }).catch(() => null)
        if (!result || result.code !== 0) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Could not confirm Windows process-tree termination')
      } else {
        try { process.kill(-pid, 'SIGTERM') } catch { terminate() }
        // A terminal parent can exit before a child that ignores SIGTERM.
        // Finish cancelling the whole process group before returning to the UI.
        await new Promise(resolve => setTimeout(resolve, 750))
        try { process.kill(-pid, 'SIGKILL') } catch { if (!finished) terminate() }
        for (let attempt = 0; attempt < 40; attempt++) {
          if (!await groupActive(pid)) break
          if (attempt === 39) throw new GoECSError('PROCESS_STOP_UNCONFIRMED', 'Local child processes did not terminate')
          await new Promise(resolve => setTimeout(resolve, 50))
        }
      }
    })()
    // A failed stop must reach the caller even if the child never exits.
    cancellation.catch(rejectStop)
    await cancellation
    return done
  }
  const timer = setTimeout(() => { onDeadline(); void stop().catch(() => {}) }, deadlineMs)
  done = Promise.race([done, stopFailure]).finally(async () => { finished = true; clearTimeout(timer); if (cancellation) await cancellation })
  return { pid, write, stop, done }
}
