import { describe, expect, it } from 'vitest'
import { describeHealth } from './health'

describe('describeHealth', () => {
  it('returns ok for a healthy backend', () => {
    expect(describeHealth({ status: 'ok' })).toBe('ok')
  })

  it('flags anything else', () => {
    expect(describeHealth({})).toBe('unexpected response')
  })
})
