"use client"

import { useEffect, useRef } from 'react'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'

interface Station {
  name: string
  location: {
    coordinates: [number, number]
  }
}

interface RouteLeg {
  from: string
  to: string
  price: number
}

interface RouteData {
  route: Station[]
  total_price: number
  legs: RouteLeg[]
}

interface MapComponentProps {
  routeData: RouteData
}

export default function MapComponent({ routeData }: MapComponentProps) {
  const mapRef = useRef<L.Map | null>(null)
  const markersRef = useRef<L.Marker[]>([])
  const containerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    // Wait for the container to be rendered
    if (!containerRef.current) return

    // Initialize map if it doesn't exist
    if (!mapRef.current) {
      mapRef.current = L.map(containerRef.current, {
        center: [9.0222, 38.7468],
        zoom: 13,
        layers: [
          L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
            attribution: '© OpenStreetMap contributors'
          })
        ]
      })

      // Ensure the map container is visible and sized correctly
      setTimeout(() => {
        if (mapRef.current) {
          mapRef.current.invalidateSize()
        }
      }, 0)
    }

    // Clear existing markers
    markersRef.current.forEach(marker => marker.remove())
    markersRef.current = []

    // Add markers for each station
    const bounds = L.latLngBounds([])
    routeData.route.forEach((station) => {
      const [lng, lat] = station.location.coordinates
      const marker = L.marker([lat, lng], {
        icon: L.divIcon({
          html: '🚖',
          className: 'taxi-marker',
          iconSize: [25, 25]
        })
      })
      .bindPopup(station.name)
      .addTo(mapRef.current!)

      markersRef.current.push(marker)
      bounds.extend([lat, lng])
    })

    // Fit map to show all markers
    if (bounds.isValid()) {
      mapRef.current?.fitBounds(bounds, { padding: [50, 50] })
    }

    // Cleanup function
    return () => {
      if (mapRef.current) {
        mapRef.current.remove()
        mapRef.current = null
      }
    }
  }, [routeData])

  return (
    <>
      <style jsx global>{`
        .taxi-marker {
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 25px;
          background: none;
          border: none;
        }
        #map-container {
          width: 100%;
          height: 60vh;
          position: relative;
          z-index: 0;
          border-radius: 0.5rem;
          overflow: hidden;
        }
        .leaflet-container {
          width: 100%;
          height: 100%;
        }
      `}</style>
      <div id="map-container" ref={containerRef} />
    </>
  )
} 