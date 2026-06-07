import { useState, useRef, useEffect, useCallback } from 'react'
import {
  format,
  addMonths,
  subMonths,
  startOfMonth,
  endOfMonth,
  eachDayOfInterval,
  startOfWeek,
  endOfWeek,
  isSameMonth,
  isSameDay,
  isToday,
  parse,
  setHours,
  setMinutes,
} from 'date-fns'
import { ptBR } from 'date-fns/locale'
import { CalendarDays, ChevronLeft, ChevronRight, Clock } from 'lucide-react'

interface Props {
  value: string
  onChange: (iso: string) => void
}

const WEEKDAYS = ['Dom', 'Seg', 'Ter', 'Qua', 'Qui', 'Sex', 'Sáb']

export function DateTimePicker({ value, onChange }: Props) {
  const [open, setOpen] = useState(false)
  const [cursorMonth, setCursorMonth] = useState(() =>
    value ? parse(value.slice(0, 7), 'yyyy-MM', new Date()) : new Date()
  )
  const [hours, setHoursState] = useState(() =>
    value ? value.slice(11, 16) : format(new Date(), 'HH:mm')
  )

  const containerRef = useRef<HTMLDivElement>(null)

  const selectedDate = value ? parse(value.slice(0, 10), 'yyyy-MM-dd', new Date()) : null

  const monthStart = startOfMonth(cursorMonth)
  const monthEnd = endOfMonth(cursorMonth)
  const calStart = startOfWeek(monthStart, { weekStartsOn: 0 })
  const calEnd = endOfWeek(monthEnd, { weekStartsOn: 0 })
  const days = eachDayOfInterval({ start: calStart, end: calEnd })

  const displayValue = selectedDate
    ? format(selectedDate, "dd 'de' MMM 'de' yyyy 'às' HH:mm", { locale: ptBR })
    : ''

  useEffect(() => {
    const handleClickOutside = (e: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [open])

  const selectDay = useCallback((day: Date) => {
    const [h, m] = hours.split(':').map(Number)
    const date = setMinutes(setHours(day, h || 0), m || 0)
    onChange(date.toISOString())
    setOpen(false)
  }, [hours, onChange])

  const prevMonth = () => setCursorMonth(prev => subMonths(prev, 1))
  const nextMonth = () => setCursorMonth(prev => addMonths(prev, 1))

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setOpen(prev => !prev)}
        className="w-full rounded-lg border border-gray-300 px-4 py-2.5 text-sm text-left flex items-center gap-2 focus:outline-none focus:ring-2 focus:ring-primary bg-white"
      >
        <CalendarDays className="h-4 w-4 text-gray-400 shrink-0" />
        <span className={displayValue ? 'text-gray-900' : 'text-gray-400'}>
          {displayValue || 'Selecionar data e hora'}
        </span>
      </button>

      {open && (
        <div className="absolute z-20 mt-1 w-[320px] rounded-xl border border-gray-200 bg-white shadow-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <button
              type="button"
              onClick={prevMonth}
              className="p-1 rounded-lg hover:bg-muted text-gray-600 transition-colors"
            >
              <ChevronLeft className="h-4 w-4" />
            </button>
            <span className="text-sm font-semibold text-gray-800">
              {format(cursorMonth, "MMMM 'de' yyyy", { locale: ptBR })}
            </span>
            <button
              type="button"
              onClick={nextMonth}
              className="p-1 rounded-lg hover:bg-muted text-gray-600 transition-colors"
            >
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>

          <div className="grid grid-cols-7 gap-0 mb-1">
            {WEEKDAYS.map(d => (
              <div key={d} className="text-center text-xs font-medium text-gray-400 py-1">
                {d}
              </div>
            ))}
          </div>

          <div className="grid grid-cols-7 gap-0">
            {days.map(day => {
              const sameMonth = isSameMonth(day, cursorMonth)
              const selected = selectedDate ? isSameDay(day, selectedDate) : false
              const today = isToday(day)
              return (
                <button
                  key={day.toISOString()}
                  type="button"
                  onClick={() => selectDay(day)}
                  className={`
                    text-sm py-1.5 rounded-lg transition-colors
                    ${!sameMonth ? 'text-gray-300' : ''}
                    ${selected ? 'bg-primary text-white font-semibold' : ''}
                    ${!selected && sameMonth && today ? 'border border-primary text-primary' : ''}
                    ${!selected && sameMonth && !today ? 'text-gray-700 hover:bg-secondary/20' : ''}
                  `}
                >
                  {format(day, 'd')}
                </button>
              )
            })}
          </div>

          <div className="mt-3 pt-3 border-t border-gray-100">
            <div className="flex items-center gap-2">
              <Clock className="h-4 w-4 text-gray-400" />
              <input
                type="time"
                value={hours}
                onChange={e => setHoursState(e.target.value)}
                className="flex-1 rounded-lg border border-gray-200 px-3 py-1.5 text-sm focus:outline-none focus:ring-2 focus:ring-primary"
              />
              {selectedDate && (
                <button
                  type="button"
                  onClick={() => selectDay(selectedDate)}
                  className="bg-primary text-white text-sm font-medium px-4 py-1.5 rounded-lg hover:brightness-110 transition-all"
                >
                  OK
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
