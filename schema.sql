CREATE TABLE {{ .table }} ({{ .columns }})

SELECT create_hypertable({{ .table|quoteLiteral }}, 'time', chunk_time_interval => INTERVAL '7d')

ALTER TABLE {{ .table }} SET (timescaledb.compress, timescaledb.compress_segmentby = 'tag_id')

CREATE TABLE
    breeze_fruit (
        "time" TIMESTAMPTZ,
        carrier_id VARCHAR(32),
        schema_ver INTEGER,
        status VARCHAR(32),
        lane INTEGER,
        frame INTEGER,
        rod INTEGER,
        cup INTEGER,
        cup_weight REAL,
        cup_wai REAL,
        side VARCHAR(32),
        batch_id INTEGER,
        batch_name VARCHAR(200),
        batch_guid UUID,
        variety_name VARCHAR(200),
        variety_guid UUID,
        product_name VARCHAR(200),
        product_guid UUID,
        product_pack VARCHAR(200),
        outlet_name VARCHAR(200),
        outlet_id INTEGER,
        outlet_totalled BOOLEAN,
        size_name VARCHAR(200),
        size_id INTEGER,
        grade_name VARCHAR(200),
        grade_id INTEGER,
        is_sampled BOOLEAN,
        area REAL,
        weight REAL,
        carton_equivalent REAL,
        density REAL,
        left_offset REAL,
        major_dim REAL,
        minor_dim REAL,
        volume REAL,
        vision_grade CHAR(3),
        vision_value REAL,
        rotation_total REAL,
        rotation_processed REAL,
        subgrade_index INTEGER,
        skin_images INTEGER,
        stem_detection_error INTEGER,
        center_offsets JSONB,
        classified_blob JSONB,
        colour JSONB,
        colour_blob JSONB,
        diameters JSONB,
        features JSONB,
        function JSONB,
        timer JSONB,
        processing_time BIGINT,
        primary_defect CHAR(3),
        other_defects CHAR(3) ARRAY,
        primary_reason VARCHAR(50),
        other_reasons VARCHAR(50) ARRAY
    )
WITH
    (
        tsdb.hypertable,
        tsdb.partition_column = 'time',
        tsdb.chunk_interval = '1d',
        tsdb.orderby = 'time DESC'
    );
CALL add_columnstore_policy('breeze_fruit', after => INTERVAL '3d');

CREATE TABLE breeze_grademap (
    id          BIGSERIAL PRIMARY KEY,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload     JSONB NOT NULL
);

CREATE INDEX breeze_grademap_received_at_idx ON breeze_grademap (received_at DESC);

CREATE TABLE breeze_grademap_changes (
    id           BIGSERIAL PRIMARY KEY,
    grademap_id  BIGINT NOT NULL REFERENCES breeze_grademap(id),
    entity_type  TEXT NOT NULL,
    entity_name  TEXT NOT NULL,
    field        TEXT NOT NULL,
    old_value    TEXT,
    new_value    TEXT NOT NULL
);

CREATE INDEX breeze_grademap_changes_grademap_id_idx ON breeze_grademap_changes (grademap_id);
CREATE INDEX breeze_grademap_changes_field_idx ON breeze_grademap_changes (field);




