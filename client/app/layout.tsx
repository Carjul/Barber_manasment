// app/layout.tsx
import './globals.css'
import { ClerkProvider } from '@clerk/nextjs'
import ReduxProvider from '../components/provider'

export const metadata = {
  title: 'Barber App',
  description: 'app para gestionar barberia',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ReduxProvider>
      <ClerkProvider>
        <html lang="es">
          <body>{children}</body>
        </html>
      </ClerkProvider>
    </ReduxProvider>
  )
}
