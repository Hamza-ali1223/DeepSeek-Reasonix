from pathlib import Path

from fontTools.pens.recordingPen import DecomposingRecordingPen
from fontTools.pens.ttGlyphPen import TTGlyphPen
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont


fonts = Path(__file__).resolve().parents[1] / "public" / "fonts"

for weight in (400, 500, 600):
    source = TTFont(fonts / "InterVariable.woff2", recalcTimestamp=False)
    instantiateVariableFont(source, {"wght": weight, "opsz": 14}, inplace=True)
    target_path = fonts / f"inter-studio-{weight}.ttf"
    target = TTFont(target_path, recalcTimestamp=False)
    glyphs = source.getGlyphSet()
    order = target.getGlyphOrder()
    for codepoint in range(0x2018, 0x2020):
        source_name = source.getBestCmap()[codepoint]
        name = target.getBestCmap().get(codepoint, f"uni{codepoint:04X}")
        recording = DecomposingRecordingPen(glyphs)
        glyphs[source_name].draw(recording)
        pen = TTGlyphPen(None)
        recording.replay(pen)
        target["glyf"][name] = pen.glyph()
        target["hmtx"][name] = source["hmtx"][source_name]
        if name not in order:
            order.append(name)
        for table in target["cmap"].tables:
            if table.isUnicode():
                table.cmap[codepoint] = name
    target.setGlyphOrder(order)
    target.save(target_path)
