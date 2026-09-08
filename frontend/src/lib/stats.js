import * as api from '../api'

// Shared by Dashboard (the full view) and Sidebar (the quick-glance
// badges) so both read the exact same numbers computed the exact same
// way — everything here comes from the same listItems/listScores
// endpoints the rest of the app already uses, nothing is stored
// separately just for these stats.
export async function computeStats(sessionsList, you) {
  const perSession = await Promise.all(
    sessionsList.map(async (session) => {
      const [items, scores] = await Promise.all([api.listItems(session.id), api.listScores(session.id)])
      return { session, items, scores }
    }),
  )

  let totalItems = 0
  let totalFlagged = 0
  let totalPending = 0
  let totalScoredByAnyone = 0
  const tasks = []
  const dimensionCounts = {}

  for (const { session, items, scores } of perSession) {
    const scoresByItem = {}
    for (const s of scores) {
      ;(scoresByItem[s.item_id] ||= []).push(s)
    }

    for (const item of items) {
      totalItems++
      const itemScores = scoresByItem[item.id] || []
      if (itemScores.length > 0) totalScoredByAnyone++

      if (item.flagged) {
        totalFlagged++
        for (const dim of item.disagreeing_dimensions || []) {
          dimensionCounts[dim] = (dimensionCounts[dim] || 0) + 1
        }
        tasks.push({
          kind: 'flagged',
          sessionId: session.id,
          sessionName: session.name,
          itemId: item.id,
          question: item.question,
          at: item.order_index,
        })
      }

      const scoredByYou = itemScores.some((s) => s.reviewer_email === you)
      if (!scoredByYou) {
        totalPending++
        tasks.push({
          kind: 'pending',
          sessionId: session.id,
          sessionName: session.name,
          itemId: item.id,
          question: item.question,
          at: item.order_index,
        })
      }
    }
  }

  // Flagged items first (more urgent), then pending; cap the list so it
  // stays a quick scan, not another whole page to read.
  tasks.sort((a, b) => (a.kind === b.kind ? 0 : a.kind === 'flagged' ? -1 : 1))

  const dimensionBars = Object.entries(dimensionCounts)
    .map(([label, value]) => ({ label, value }))
    .sort((a, b) => b.value - a.value)

  return {
    sessionCount: sessionsList.length,
    totalItems,
    totalFlagged,
    totalPending,
    scoredPercent: totalItems === 0 ? 0 : (totalScoredByAnyone / totalItems) * 100,
    agreementPercent: totalItems === 0 ? 100 : ((totalItems - totalFlagged) / totalItems) * 100,
    tasks: tasks.slice(0, 8),
    dimensionBars,
  }
}
