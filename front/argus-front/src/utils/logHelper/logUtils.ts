/**
 * Time:2024/8/26 下午5:02 17
 * Name:logUtils.ts
 * Path:src/utils
 * ProjectName:utopia-front-vue
 * Author:charlatans
 *
 *  Il n'ya qu'un héroïsme au monde :
 *     c'est de voir le monde tel qu'il est et de l'aimer.
 */
import app from '@/constants/app'
import type { Color } from '@/constants/logUtil'
import { COLOR_MAP } from '@/constants/logUtil'
import {
  LevelName,
  LogLevelEnum
} from '@/utils/logHelper/logEnum'

const env = import.meta.env
// 是否输出日志
const isPrint = env.VITE_MODE !== app.PRODUCTION
const COLORS: Color[] = [
  'primary',
  'success',
  'info',
  'warning',
  'danger',
  'error'
]

// 日志输出等级
const logPrintLevel = LogLevelEnum.DEBUG

const getColor = (type: Color) => COLOR_MAP[type]

/**
 * 输出函数
 */
type Log<T extends any[]> = (type: Color, ...args: T) => void

/**
 * 返回值
 */
type Created<T extends any[]> = Record<Color, (...args: T) => void>

const createLog = <T extends any[]> (
  fn: Log<T>
): Created<T> => {
  return COLORS.reduce((logs, type) => {
    logs[type] = (...args: T) => fn(type, ...args)
    return logs
  }, {} as Created<T>)
}

const nsLog = (type: Color, ns: string, msg: string, ...args: any[]) => {
  const color = getColor(type)
  console.log(
    `%c ${ns} %c ${msg} %c ${args.length ? '%o' : ''}`,
    `background:${color};border:1px solid ${color}; padding: 1px; border-radius: 4px 0 0 4px; color: #fff;`,
    `border:1px solid ${color}; padding: 1px; border-radius: 0 4px 4px 0; color: ${color};`,
    'background:transparent',
    ...args
  )
}

/**
 * 打印格式化日志
 */
export const logN = createLog(nsLog)

const sLog = (type: Color, msg: string, ...args: any[]) => {
  const color = getColor(type)
  console.log(
    `%c ${msg} ${args.length ? '%o' : ''}`,
    `color: ${color};`,
    ...args
  )
}

/**
 * 打印带有颜色的日志
 */
export const logS = createLog(sLog)

const bLog = (type: Color, msg: string, ...args: any[]) => {
  const color = getColor(type)
  console.log(
    `%c ${msg} ${args.length ? '%o' : ''}`,
    `background:${color}; padding: 2px; border-radius: 4px; color: #fff;`,
    ...args
  )
}

/**
 * 打印带有颜色背景的日志
 */
export const logB = createLog(bLog)

export const disLog = () => {
  console.log = function() {}
}

const logHelper = {
  /**
   * DEBUG
   */
  d: (tag: string, msg: string) => {

  }
}

const logStyleFormat = (): string => {
  return ''
}

/**
 * 打印日志
 * @param level 日志等级
 * @param tag 标记
 * @param msg 信息
 * @param e 报错信息【存在报错信息，默认使用 **ERROR** 等级】
 */
const print = (
  level: LogLevelEnum, tag: string, msg: string,
  e: Error | null | undefined = null) => {
  if (logPrintLevel > level) {
    return
  }
  // 格式化字符串
  // let str = formatMessage(msg)
  //
  // // 获取日志等级名称
  // let levelName = getLogLevelName(level)

  switch (level) {
    case LogLevelEnum.VERBOSE:
      console.log()
      return LevelName.VERBOSE
    case LogLevelEnum.DEBUG:
      return LevelName.DEBUG
    case LogLevelEnum.INFO:
      return LevelName.INFO
    case LogLevelEnum.WARN:
      return LevelName.WARN
    case LogLevelEnum.ERROR:
      return LevelName.ERROR
  }
}

/**
 * 格式化输出信息
 * @param str
 */
const formatMessage = (str: string): string => {
  return `[${str}]`
}
