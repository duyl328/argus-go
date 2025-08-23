declare module 'justified-layout' {
  interface GeometryOptions {
    containerWidth?: number
    targetRowHeight?: number
    targetRowHeightTolerance?: number
    boxSpacing?: number
    containerPadding?: number
  }

  interface Box {
    top: number
    left: number
    width: number
    height: number
  }

  interface Geometry {
    containerHeight: number
    boxes: Box[]
  }

  interface PhotoInput {
    width: number
    height: number
  }

  function justifiedLayout(photos: PhotoInput[], options?: GeometryOptions): Geometry
  export = justifiedLayout
}