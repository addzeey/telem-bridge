/* eslint-disable */

// @ts-nocheck

// noinspection JSUnusedGlobalSymbols

// This file was automatically generated by TanStack Router.
// You should NOT make any changes in this file as it will be overwritten.
// Additionally, you should also exclude this file from your linter and/or formatter to prevent it from being checked or modified.

// Import Routes

import { Route as rootRoute } from './routes/__root'
import { Route as IndexImport } from './routes/index'
import { Route as SettingsIndexImport } from './routes/settings/index'
import { Route as TelemetryLiveImport } from './routes/telemetry/live'
import { Route as SettingsTelemetryImport } from './routes/settings/telemetry'
import { Route as SettingsOscImport } from './routes/settings/osc'

// Create/Update Routes

const IndexRoute = IndexImport.update({
  id: '/',
  path: '/',
  getParentRoute: () => rootRoute,
} as any)

const SettingsIndexRoute = SettingsIndexImport.update({
  id: '/settings/',
  path: '/settings/',
  getParentRoute: () => rootRoute,
} as any)

const TelemetryLiveRoute = TelemetryLiveImport.update({
  id: '/telemetry/live',
  path: '/telemetry/live',
  getParentRoute: () => rootRoute,
} as any)

const SettingsTelemetryRoute = SettingsTelemetryImport.update({
  id: '/settings/telemetry',
  path: '/settings/telemetry',
  getParentRoute: () => rootRoute,
} as any)

const SettingsOscRoute = SettingsOscImport.update({
  id: '/settings/osc',
  path: '/settings/osc',
  getParentRoute: () => rootRoute,
} as any)

// Populate the FileRoutesByPath interface

declare module '@tanstack/react-router' {
  interface FileRoutesByPath {
    '/': {
      id: '/'
      path: '/'
      fullPath: '/'
      preLoaderRoute: typeof IndexImport
      parentRoute: typeof rootRoute
    }
    '/settings/osc': {
      id: '/settings/osc'
      path: '/settings/osc'
      fullPath: '/settings/osc'
      preLoaderRoute: typeof SettingsOscImport
      parentRoute: typeof rootRoute
    }
    '/settings/telemetry': {
      id: '/settings/telemetry'
      path: '/settings/telemetry'
      fullPath: '/settings/telemetry'
      preLoaderRoute: typeof SettingsTelemetryImport
      parentRoute: typeof rootRoute
    }
    '/telemetry/live': {
      id: '/telemetry/live'
      path: '/telemetry/live'
      fullPath: '/telemetry/live'
      preLoaderRoute: typeof TelemetryLiveImport
      parentRoute: typeof rootRoute
    }
    '/settings/': {
      id: '/settings/'
      path: '/settings'
      fullPath: '/settings'
      preLoaderRoute: typeof SettingsIndexImport
      parentRoute: typeof rootRoute
    }
  }
}

// Create and export the route tree

export interface FileRoutesByFullPath {
  '/': typeof IndexRoute
  '/settings/osc': typeof SettingsOscRoute
  '/settings/telemetry': typeof SettingsTelemetryRoute
  '/telemetry/live': typeof TelemetryLiveRoute
  '/settings': typeof SettingsIndexRoute
}

export interface FileRoutesByTo {
  '/': typeof IndexRoute
  '/settings/osc': typeof SettingsOscRoute
  '/settings/telemetry': typeof SettingsTelemetryRoute
  '/telemetry/live': typeof TelemetryLiveRoute
  '/settings': typeof SettingsIndexRoute
}

export interface FileRoutesById {
  __root__: typeof rootRoute
  '/': typeof IndexRoute
  '/settings/osc': typeof SettingsOscRoute
  '/settings/telemetry': typeof SettingsTelemetryRoute
  '/telemetry/live': typeof TelemetryLiveRoute
  '/settings/': typeof SettingsIndexRoute
}

export interface FileRouteTypes {
  fileRoutesByFullPath: FileRoutesByFullPath
  fullPaths:
    | '/'
    | '/settings/osc'
    | '/settings/telemetry'
    | '/telemetry/live'
    | '/settings'
  fileRoutesByTo: FileRoutesByTo
  to:
    | '/'
    | '/settings/osc'
    | '/settings/telemetry'
    | '/telemetry/live'
    | '/settings'
  id:
    | '__root__'
    | '/'
    | '/settings/osc'
    | '/settings/telemetry'
    | '/telemetry/live'
    | '/settings/'
  fileRoutesById: FileRoutesById
}

export interface RootRouteChildren {
  IndexRoute: typeof IndexRoute
  SettingsOscRoute: typeof SettingsOscRoute
  SettingsTelemetryRoute: typeof SettingsTelemetryRoute
  TelemetryLiveRoute: typeof TelemetryLiveRoute
  SettingsIndexRoute: typeof SettingsIndexRoute
}

const rootRouteChildren: RootRouteChildren = {
  IndexRoute: IndexRoute,
  SettingsOscRoute: SettingsOscRoute,
  SettingsTelemetryRoute: SettingsTelemetryRoute,
  TelemetryLiveRoute: TelemetryLiveRoute,
  SettingsIndexRoute: SettingsIndexRoute,
}

export const routeTree = rootRoute
  ._addFileChildren(rootRouteChildren)
  ._addFileTypes<FileRouteTypes>()

/* ROUTE_MANIFEST_START
{
  "routes": {
    "__root__": {
      "filePath": "__root.tsx",
      "children": [
        "/",
        "/settings/osc",
        "/settings/telemetry",
        "/telemetry/live",
        "/settings/"
      ]
    },
    "/": {
      "filePath": "index.tsx"
    },
    "/settings/osc": {
      "filePath": "settings/osc.tsx"
    },
    "/settings/telemetry": {
      "filePath": "settings/telemetry.tsx"
    },
    "/telemetry/live": {
      "filePath": "telemetry/live.tsx"
    },
    "/settings/": {
      "filePath": "settings/index.tsx"
    }
  }
}
ROUTE_MANIFEST_END */
