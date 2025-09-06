
type RightViewProps = {
    selected: string;
}

export function RightView({ selected }: RightViewProps) {
    return (
        <div className="split-pane right-pane">
            {selected ? (
                <h2>{selected}</h2>
            ) : (
                <h2>Right Pane</h2>
            )}
        </div>
    )
}
