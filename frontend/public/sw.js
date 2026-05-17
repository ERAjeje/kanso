self.addEventListener('push', event => {
  const data = event.data?.json() ?? { title: 'Kanso', body: 'Como você está se sentindo agora?' }
  event.waitUntil(
    self.registration.showNotification(data.title, {
      body: data.body,
      icon: '/icon-192.png',
      badge: '/badge-72.png',
      vibrate: [200, 100, 200],
      data: { url: '/register' },
    })
  )
})

self.addEventListener('notificationclick', event => {
  event.notification.close()
  event.waitUntil(
    clients.matchAll({ type: 'window', includeUncontrolled: true }).then(clientList => {
      if (clientList.length > 0) {
        clientList[0].focus()
        clientList[0].navigate(event.notification.data?.url || '/register')
      } else {
        clients.openWindow(event.notification.data?.url || '/register')
      }
    })
  )
})
